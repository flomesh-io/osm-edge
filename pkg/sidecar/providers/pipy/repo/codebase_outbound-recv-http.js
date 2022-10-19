// version: '2022.10.09'
((
  {
    config,
    debugLogLevel,
    namespace,
    kind,
    name,
    pod,
    metrics,
    outClustersConfigs,
    specEnableEgress,
    codeMessage
  } = pipy.solve('config.js')) => (

  pipy({
  })

    .import({
      logZipkin: 'main',
      logLogging: 'main',
      _outMatch: 'main',
      _outTarget: 'main',
      _egressMode: 'main',
      _outSourceCert: 'main',
      _outRequestTime: 'main',
      _outBytesStruct: 'main',
      _outLoggingData: 'main',
      _outZipkinData: 'main',
      _outIP: 'main',
      _outPort: 'main',
      _egressTargetMap: 'main',
      _upstreamClusterName: 'main'
    })

    //
    // Analyze outbound HTTP request headers and match routes
    //
    .pipeline()
    .handleMessageStart(
      (msg) => (
        ((service, route, match, target, headers) => (
          headers = msg.head.headers,

          service = _outMatch.HttpHostPort2Service?.[headers.host],

          // Find route by HTTP host
          route = service && _outMatch.HttpServiceRouteRules?.[service],

          // Find a match by the service's route rules
          match = route?.find?.(o => (
            // Match methods
            (!o.Methods || o.Methods[msg.head.method]) &&
            // Match service whitelist
            (!o.AllowedServices || o.AllowedServices[headers.serviceidentity]) &&
            // Match path pattern
            o.Path.test(msg.head.path) &&
            // Match headers
            (!o.Headers || o.Headers.every(([k, v]) => v.test(headers[k] || '')))
          )),

          // Layer 7 load balance
          _outTarget = ((index, connIdx = null) => (
            _upstreamClusterName = match?.TargetClusters?.next?.()?.id,

            // Limit for connection
            outClustersConfigs?.[_upstreamClusterName]?.RateLimit && (
              (index = outClustersConfigs[_upstreamClusterName].RateLimit.next()) && (connIdx = outClustersConfigs[_upstreamClusterName].RateLimitObject[index.id])
            ),

            // Egress mTLS certs
            _outSourceCert = outClustersConfigs?.[_upstreamClusterName]?.SourceCert,

            outClustersConfigs?.[
              _upstreamClusterName
            ]?.Endpoints?.next?.(connIdx)
          ))(),

          // no HttpHostPort2Service
          _outMatch && !service && console.log(codeMessage('NoService'), headers?.host),

          // no TargetClusters
          match && service && !_upstreamClusterName && console.log(codeMessage('NoRoute'), service),

          // no ClustersConfigs
          match && _upstreamClusterName && !_outTarget && console.log(codeMessage('NoEndpoint'), _upstreamClusterName),

          // Add serviceidentity for request authentication
          _outTarget && (headers['serviceidentity'] = _outMatch.ServiceIdentity),

          // Add x-b3 tracing Headers
          _outTarget && metrics.funcTracingHeaders(namespace, kind, name, pod, headers, _outMatch?.Protocol),

          // Initialize ZipKin tracing data
          logZipkin && (_outZipkinData = metrics.funcMakeZipKinData(name, msg, headers, _upstreamClusterName, 'CLIENT', false)),

          // Initialize Outbound logging data
          logLogging && (_outLoggingData = {
            reqTime: Date.now(),
            meshName: os.env.MESH_NAME || '',
            remoteAddr: __inbound?.destinationAddress,
            remotePort: __inbound?.destinationPort,
            localAddr: __inbound?.remoteAddress,
            localPort: __inbound?.remotePort,
            node: {
              ip: os.env.POD_IP || '127.0.0.1',
              name: os.env.HOSTNAME || 'localhost',
            },
            pod: {
              ns: os.env.POD_NAMESPACE || 'default',
              ip: os.env.POD_IP || '127.0.0.1',
              name: os.env.POD_NAME || os.env.HOSTNAME || 'localhost',
            },
            service: {
              name: service || 'anonymous', target: _outTarget?.id, egressMode: Boolean(_egressMode)
            },
            trace: {
              id: headers?.['x-b3-traceid'] || '',
              span: headers?.['x-b3-spanid'] || '',
              parent: headers?.['x-b3-parentspanid'] || '',
              sampled: headers?.['x-b3-sampled'] || ''
            }
          }),

          _outBytesStruct = {},
          _outBytesStruct.requestSize = _outBytesStruct.responseSize = 0,

          // EGRESS mode
          !_outTarget && (specEnableEgress || _outMatch?.AllowedEgressTraffic) && (
            target = _outIP + ':' + _outPort,
            _upstreamClusterName = target,
            !_egressTargetMap[target] && (_egressTargetMap[target] = new algo.RoundRobinLoadBalancer({
              [target]: 100
            })),
            _outTarget = _egressTargetMap[target].next(),
            _egressMode = true
          ),

          _outRequestTime = Date.now(),
          msg.head && (msg.head._upstreamClusterName = _upstreamClusterName),

          debugLogLevel && (
            console.log('outbound path: ', msg.head.path),
            console.log('outbound headers: ', msg.head.headers),
            console.log('outbound service: ', service),
            console.log('outbound route: ', route),
            console.log('outbound match: ', match),
            console.log('outbound _outTarget: ', _outTarget?.id)
          )
        ))()
      )
    )
    .handleMessage(
      msg => (
        logLogging && (
          _outLoggingData.req = Object.assign({}, msg.head),
          _outLoggingData.req['body'] = msg.body.toString('base64'),
          _outLoggingData['reqSize'] = msg.body.size
        )
      )
    )
    .branch(
      () => config?.outClustersBreakers?.[_upstreamClusterName]?.block?.(), $ => $
        .replaceMessage(
          () => config.outClustersBreakers[_upstreamClusterName].message()
        ),
      () => Boolean(_outTarget), $ => $
        .chain(['outbound-mux-http.js', 'outbound-breaker.js']),
      $ => $
        .replaceMessage(
          new Message({
            status: 403
          }, 'Access denied')
        )
    )

))()