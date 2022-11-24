(
  (
    {
      namespace,
      pod,
      allowedEndpoints,
      inTrafficMatches,
    } = pipy.solve('config.js'),

    metrics = pipy.solve('metrics-init.js'),

    {
      debug,
    } = pipy.solve('utils.js'),

    connLimitCache = new algo.Cache((portMatches) => (
      (portMatches?.RateLimit?.Local?.Connections > 0) ? (
        {
          ConnectionLimitQuota: new algo.Quota(
            portMatches?.RateLimit?.Local?.Burst ? portMatches.RateLimit.Local.Burst : portMatches.RateLimit.Local.Connections,
            {
              produce: portMatches.RateLimit.Local.Connections,
              per: portMatches?.RateLimit?.Local?.StatTimeWindow > 0 ? portMatches.RateLimit.Local.StatTimeWindow : 1
            }
          ),
          ConnectionLimitStatsKey: 'local_rate_limit.inbound_' + namespace + '/' + pod.split('-')[0] + '_' + portMatches.Port + '_' + portMatches.Protocol + '.rate_limited'
        }
      ) : null
    ), null, {})

  ) => (

    pipy({
    })

      .import({
        _flow: 'main',
      })

      .export('inbound-classifier', {
        _inMatch: null,
        _inTarget: null,
        _ingressEnable: false,
        _localClusterName: null,
      })

      .pipeline()

      .replaceStreamStart(
        (evt) => (
          _flow = 'inbound',

          ((remoteAddress = __inbound.remoteAddress || '127.0.0.1', connLimit = null) => (

            // Find a match by destination port
            _inMatch = (
              allowedEndpoints?.[remoteAddress] &&
              inTrafficMatches?.[__inbound.destinationPort || 0]
            ),

            // Check client address against the whitelist
            _inMatch?.AllowedEndpoints && (_inMatch.AllowedEndpoints[remoteAddress] === undefined) && (
              _inMatch = null
            ),

            // Check RateLimit.Local.Connections
            _inMatch && (connLimit = connLimitCache.get(_inMatch)) && (
              (connLimit.ConnectionLimitQuota && connLimit.ConnectionLimitQuota.consume(1) != 1) && (
                metrics.sidecarInsideStats[connLimit.ConnectionLimitStatsKey] += 1,
                _inMatch = null
              )
            ),

            debug(log => (
              log('inbound _inMatch: ', _inMatch),
              log('inbound protocol: ', _inMatch?.Protocol)
            ))
          ))(),
          !_inMatch || (_inMatch.Protocol !== 'http' && _inMatch.Protocol !== 'grpc') ? evt : null
        )
      )

      .branch(
        () => (_inMatch?.Protocol === 'http' || _inMatch?.Protocol === 'grpc'), $ => $
          .chain([
            'inbound-tls-termination.js',
            'inbound-demux-http.js',
            'inbound-http-routing.js',
            'metrics-http.js',
            'tracing.js',
            'logging.js',
            'inbound-throttle.js',
            'inbound-mux-http.js',
            'metrics-tcp.js',
            'inbound-proxy-tcp.js'
          ]),

        () => Boolean(_inMatch), $ => $
          .chain([
            'inbound-tls-termination.js',
            'inbound-tcp-load-balance.js',
            'metrics-tcp.js',
            'inbound-proxy-tcp.js'
          ]),

        $ => $
          .replaceStreamStart(
            new StreamEnd('ConnectionReset')
          )
      )

  ))()
