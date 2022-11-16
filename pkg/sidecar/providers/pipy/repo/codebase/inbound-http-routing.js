((
  {
    inClustersConfigs,
  } = pipy.solve('config.js'),
  {
    debug,
    shuffle,
  } = pipy.solve('utils.js'),

  routeRulesCache = new algo.Cache((routeRules) => (
    routeRules ? Object.entries(routeRules).map(
      ([path, condition]) => (
        {
          RuleName: path,
          Path: new RegExp(path),
          Methods: condition.Methods && Object.fromEntries(condition.Methods.map(e => [e, true])),
          Headers: condition.Headers && Object.entries(condition.Headers).map(([k, v]) => [k, new RegExp(v)]),
          AllowedServices: condition.AllowedServices && Object.fromEntries(condition.AllowedServices.map(e => [e, true])),
          TargetClusters: condition.TargetClusters && new algo.RoundRobinLoadBalancer(shuffle(condition.TargetClusters))
        }
      )
    ) : null
  ), null, {}),

  clustersConfigCache = new algo.Cache((clustersConfig) => (
    clustersConfig ? new algo.RoundRobinLoadBalancer(shuffle(clustersConfig)) : null
  ), null, {}),

  connLimitCache = new algo.Cache((connLimit) => (
    connLimit?.Local?.Connections > 0 ? ((array) => (
      array = (new Array(connLimit?.Local?.Connections > 1000 ? 1000 : connLimit?.Local?.Connections)).fill('connection_').map((e, index) => e + index),
      {
        ConnectionLimit: new algo.LeastWorkLoadBalancer(Object.fromEntries(array.map(k => [k, 1]))),
        ConnectionLimitObject: Object.fromEntries(array.map(k => [k, new String(k)])),
      }
    ))() : null
  ), null, {})

) => (

  pipy({
    _inSessionClose: null
  })

    .import({
      _inMatch: 'inbound-classifier',
      _inTarget: 'inbound-classifier',
      _ingressEnable: 'inbound-classifier',
      _localClusterName: 'inbound-classifier',
    })

    .export('inbound-http-routing', {
      _inService: null,
      _inRouteRule: null,
    })

    //
    // Analyze inbound HTTP request headers and match routes
    //
    .pipeline()

    .handleMessageStart(
      (msg) => (
        ((routeRules, matchRoute, connLimit, headers, connIdx = null) => (
          headers = msg.head.headers,

          // INGRESS mode
          // When found in SourceIPRanges, service is '*'
          _ingressEnable && (_inService = '*'),

          // Find the service
          // When serviceidentity is present, service is headers.host
          !_inService && (_inService = (headers.serviceidentity && _inMatch?.HttpHostPort2Service?.[headers.host])),

          _inMatch.HttpServiceRouteRules?.[_inService]?.RouteRules && (
            routeRules = routeRulesCache.get(_inMatch.HttpServiceRouteRules[_inService].RouteRules)
          ),

          // Find a match by the service's route rules
          matchRoute = routeRules?.find?.(o => (
            // Match methods
            (!o.Methods || o.Methods[msg.head.method]) &&
            // Match service whitelist
            (!o.AllowedServices || o.AllowedServices[headers.serviceidentity]) &&
            // Match path pattern
            o.Path.test(msg.head.path) &&
            // Match headers
            (!o.Headers || o.Headers.every(([k, v]) => v.test(headers[k] || '')))
          )),

          matchRoute && (

            _inRouteRule = matchRoute.RuleName,

            // Limit for connection
            (connLimit = connLimitCache.get(_inMatch?.RateLimit)) && ((index) => (
              (index = connLimit.ConnectionLimit.next()) && (connIdx = connLimit.ConnectionLimitObject[index.id])
            ))(),

            // Layer 7 load balance
            _inTarget = (
              _localClusterName = matchRoute?.TargetClusters?.next?.()?.id,
              clustersConfigCache.get(inClustersConfigs?.[_localClusterName])?.next?.(connIdx)
            )
          ),

          // Close sessions from any HTTP proxies
          !_inTarget && headers['x-forwarded-for'] && (
            _inSessionClose = true
          ),

          debug(log => (
            log('inbound path: ', msg.head.path),
            log('inbound headers: ', msg.head.headers),
            log('inbound _inService: ', _inService),
            log('inbound _inRouteRule: ', _inRouteRule),
            log('inbound _localClusterName: ', _localClusterName),
            log('inbound _inTarget: ', _inTarget?.id)
          ))
        ))()
      )
    )

    .branch(
      () => Boolean(_inTarget), $ => $
        .chain(),

      () => _inSessionClose, $ => $
        .replaceMessageStart(
          new StreamEnd('ConnectionRefused')
        ),

      $ => $
        .replaceMessage(
          new Message({
            status: 403
          }, 'Access denied')
        )
    )

))()