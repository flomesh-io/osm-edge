((
  {
    outClustersConfigs,
    specEnableEgress,
    tlsCertChain,
    tlsPrivateKey,
    addIssuingCA,
  } = pipy.solve('config.js'),
  codeMessage = pipy.solve('codes.js'),
  {
    debug,
    shuffle,
  } = pipy.solve('utils.js'),

  routeRulesCache = new algo.Cache((routeRules) => (
    routeRules ? Object.entries(routeRules).map(
      ([path, condition]) => ({
        RuleName: path,
        Path: new RegExp(path),
        Methods: condition.Methods && Object.fromEntries(condition.Methods.map(e => [e, true])),
        Headers: condition.Headers && Object.entries(condition.Headers).map(([k, v]) => [k, new RegExp(v)]),
        AllowedServices: condition.AllowedServices && Object.fromEntries(condition.AllowedServices.map(e => [e, true])),
        TargetClusters: condition.TargetClusters && new algo.RoundRobinLoadBalancer(shuffle(condition.TargetClusters))
      })
    ) : null
  ), null, {}),

  clustersConfigCache = new algo.Cache((clustersConfig) => (
    ((obj = {}, array) => (
      clustersConfig?.ConnectionSettings?.tcp?.MaxConnections && (
        array = (new Array(clustersConfig.ConnectionSettings?.tcp?.MaxConnections > 1000 ? 1000 : clustersConfig.ConnectionSettings?.tcp?.MaxConnections)).fill('connection_').map((e, index) => e + index),
        obj['ConnectionLimit'] = new algo.LeastWorkLoadBalancer(Object.fromEntries(array.map(k => [k, 1]))),
        obj['ConnectionLimitObject'] = Object.fromEntries(array.map(k => [k, new String(k)]))
      ),
      obj['Endpoints'] = new algo.RoundRobinLoadBalancer(shuffle(clustersConfig.Endpoints)),
      clustersConfig?.SourceCert?.CertChain && clustersConfig?.SourceCert?.PrivateKey && clustersConfig?.SourceCert?.IssuingCA && (
        obj['SourceCert'] = clustersConfig.SourceCert,
        addIssuingCA(clustersConfig.SourceCert.IssuingCA)
      ),
      clustersConfig?.SourceCert?.OsmIssued && tlsCertChain && tlsPrivateKey && (
        obj['SourceCert'] = { CertChain: tlsCertChain, PrivateKey: tlsPrivateKey }
      ),
      obj['JsonObject'] = clustersConfig,
      obj
    ))()
  ), null, {})

) => (

  pipy({
    egressTargetMap: {}
  })

    .import({
      _outIP: 'outbound-classifier',
      _outPort: 'outbound-classifier',
      _outMatch: 'outbound-classifier',
      _outTarget: 'outbound-classifier',
      _egressEnable: 'outbound-classifier',
      _outSourceCert: 'outbound-classifier',
      _upstreamClusterName: 'outbound-classifier'
    })

    .export('outbound-http-routing', {
      _outService: null,
      _outRouteRule: null,
    })

    //
    // Analyze outbound HTTP request headers and match routes
    //
    .pipeline()

    .handleMessageStart(
      (msg) => (
        ((route, match, target, headers) => (
          headers = msg.head.headers,

          _outService = _outMatch.HttpHostPort2Service?.[headers.host],

          // Find route by HTTP host
          route = _outService && routeRulesCache.get(_outMatch.HttpServiceRouteRules?.[_outService]),

          // Find a match by the _outService's route rules
          match = route?.find?.(o => (
            // Match methods
            (!o.Methods || o.Methods[msg.head.method]) &&
            // Match _outService whitelist
            (!o.AllowedServices || o.AllowedServices[headers.serviceidentity]) &&
            // Match path pattern
            o.Path.test(msg.head.path) &&
            // Match headers
            (!o.Headers || o.Headers.every(([k, v]) => v.test(headers[k] || '')))
          )),

          match && (
            _outRouteRule = match.RuleName,

            // Layer 7 load balance
            _outTarget = ((clusterConfig, index, connIdx = null) => (
              _upstreamClusterName = match?.TargetClusters?.next?.()?.id,

              // Limit for connection
              clusterConfig = clustersConfigCache.get(outClustersConfigs?.[_upstreamClusterName]),

              clusterConfig?.ConnectionLimit && (
                (index = clusterConfig.ConnectionLimit.next()) && (connIdx = clusterConfig.ConnectionLimitObject[index.id])
              ),

              // Egress mTLS certs
              _outSourceCert = clusterConfig?.SourceCert,

              clusterConfig?.Endpoints?.next?.(connIdx)
            ))()

          ),

          // no HttpHostPort2Service
          _outMatch && !_outService && console.log(codeMessage('NoService'), headers?.host),

          // no TargetClusters
          match && _outService && !_upstreamClusterName && console.log(codeMessage('NoRoute'), _outService),

          // no ClustersConfigs
          match && _upstreamClusterName && !_outTarget && console.log(codeMessage('NoEndpoint'), _upstreamClusterName),

          // Add serviceidentity for request authentication
          _outTarget && (headers['serviceidentity'] = _outMatch.ServiceIdentity),

          // EGRESS mode
          !_outTarget && (specEnableEgress || _outMatch?.AllowedEgressTraffic) && (
            target = _outIP + ':' + _outPort,
            _upstreamClusterName = target,
            !egressTargetMap[target] && (egressTargetMap[target] = new algo.RoundRobinLoadBalancer({
              [target]: 100
            })),
            _outTarget = egressTargetMap[target].next(),
            _egressEnable = true
          ),

          debug(log => (
            log('outbound path: ', msg.head.path),
            log('outbound headers: ', msg.head.headers),
            log('outbound _outService: ', _outService),
            log('outbound _outRouteRule: ', _outRouteRule),
            log('outbound _upstreamClusterName: ', _upstreamClusterName),
            log('outbound _outTarget: ', _outTarget?.id)
          ))
        ))()
      )
    )

    .branch(
      () => Boolean(_outTarget), $ => $
        .chain(),

      $ => $
        .replaceMessage(
          new Message({
            status: 403
          }, 'Access denied')
        )
    )

))()