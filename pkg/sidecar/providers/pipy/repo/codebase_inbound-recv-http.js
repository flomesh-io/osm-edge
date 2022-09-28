// version: '2022.09.28'
((
  {
    name,
    metrics,
    debugLogLevel,
    inClustersConfigs
  } = pipy.solve('config.js')) => (

  pipy({
    _inRateLimit: null
  })

    .import({
      logZipkin: 'main',
      logLogging: 'main',
      _inMatch: 'main',
      _inTarget: 'main',
      _ingressMode: 'main',
      _inBytesStruct: 'main',
      _inLoggingData: 'main',
      _inZipkinData: 'main',
      _inSessionControl: 'main',
      _localClusterName: 'main',
      _forbiddenTLS: 'main'
    })

    .export('inbound-recv-http', {
      _inHostRateLimit: null,
      _inPathRateLimit: null,
      _inHeaderRateLimit: null,
    })

    //
    // Analyze inbound HTTP request headers and match routes
    //
    .pipeline()
    .handleMessageStart(
      (msg) => (
        ((service, match, headers, connIdx) => (
          headers = msg.head.headers,

          // INGRESS mode
          // When found in SourceIPRanges, service is '*'
          _ingressMode && (service = '*'),

          // Find the service
          // When serviceidentity is present, service is headers.host
          !service && (service = (headers.serviceidentity && _inMatch?.HttpHostPort2Service?.[headers.host])),

          // Find a match by the service's route rules
          match = !Boolean(_forbiddenTLS) && _inMatch.HttpServiceRouteRules?.[service]?.RouteRules?.find?.(o => (
            // Match methods
            (!o.Methods || o.Methods[msg.head.method]) &&
            // Match service whitelist
            (!o.AllowedServices || o.AllowedServices[headers.serviceidentity]) &&
            // Match path pattern
            o.Path.test(msg.head.path) &&
            // Match headers
            (!o.Headers || o.Headers.every(([k, v]) => v.test(headers[k] || '')))
          )),

          // Limit for connection
          _inMatch?.RateLimit && ((index) => (
            (index = _inMatch.RateLimit.next()) && (connIdx = _inMatch.RateLimitObject[index.id])
          ))(),

          // Layer 7 load balance
          _inTarget = (
            inClustersConfigs[
              _localClusterName = match?.TargetClusters?.next?.()?.id
            ]?.next?.(connIdx)
          ),

          // Inbound rate limit quotas.
          _inTarget && ((rt) => (
            _inHostRateLimit = _inMatch.HttpServiceRouteRules?.[service]?.RateLimit,
            _inPathRateLimit = match?.RateLimit,
            rt = _inMatch.HttpServiceRouteRules?.[service]?.HeaderRateLimits?.find?.(o => (
              (!o.Headers || o.Headers.every(([k, v]) => v.test(headers[k] || ''))))),
            _inHeaderRateLimit = rt?.RateLimit
          ))(),

          // Close sessions from any HTTP proxies
          !_inTarget && headers['x-forwarded-for'] && (
            _inSessionControl.close = true
          ),

          // Initialize ZipKin tracing data
          logZipkin && (_inZipkinData = metrics.funcMakeZipKinData(name, msg, headers, _localClusterName, 'SERVER', true)),

          // Initialize Inbound logging data
          logLogging && (_inLoggingData = metrics.funcMakeLoggingData(msg, 'inbound')),

          _inBytesStruct = {},
          _inBytesStruct.requestSize = _inBytesStruct.responseSize = 0,

          debugLogLevel && (
            console.log('inbound path: ', msg.head.path),
            console.log('inbound headers: ', msg.head.headers),
            console.log('inbound service: ', service),
            console.log('inbound match: ', match),
            console.log('inbound _inTarget: ', _inTarget?.id)
          )
        ))()
      )
    )
    .chain(['inbound-throttle.js'])
    .handleMessageStart(
      msg => Boolean(msg.head?.overflow) && (_inRateLimit = msg.head?.ratelimit)
    )

    .branch(
      () => _inRateLimit, $ => $
        .replaceMessage(
          () => (
            ((hds = {}) => (
              metrics.sidecarInsideStats['http_local_rate_limiter.http_local_rate_limit.rate_limited'] += 1,
              _inRateLimit?.headers && _inRateLimit.headers.forEach(h => hds[h.Name] = h.Value),
              new Message({
                status: _inRateLimit.status,
                headers: hds
              }, 'Too Many Requests')
            ))()
          )
        ),
      () => Boolean(_inTarget) && _inMatch?.Protocol === 'grpc', $ => $
        .muxHTTP(() => _inTarget, {
          version: 2
        }).to($ => $.chain(['inbound-proxy-tcp.js'])),
      () => Boolean(_inTarget), $ => $
        .muxHTTP(() => _inTarget).to($ => $.chain(['inbound-proxy-tcp.js'])),
      $ => $
        .replaceMessage(
          new Message({
            status: 403
          }, 'Access denied')
        )
    )
    .chain()

))()