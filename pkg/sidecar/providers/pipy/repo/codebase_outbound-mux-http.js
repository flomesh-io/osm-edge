((
  {
    outClustersConfigs,
  } = pipy.solve('config.js'),
  metrics = pipy.solve('metrics-init.js'),

  clustersConfigCache = new algo.Cache((clustersConfig) => (
    ((obj = {}) => (
      clustersConfig?.RetryPolicy?.NumRetries && (
        obj['RetryPolicy'] = {
          RetryOn: clustersConfig.RetryPolicy?.RetryOn,
          lowerbound: clustersConfig.RetryPolicy?.RetryOn ? clustersConfig.RetryPolicy.RetryOn.replaceAll('x', '0') : 500,
          upperbound: clustersConfig.RetryPolicy?.RetryOn ? clustersConfig.RetryPolicy.RetryOn.replaceAll('x', '9') : 599,
          PerTryTimeout: clustersConfig.RetryPolicy?.PerTryTimeout ? clustersConfig.RetryPolicy.PerTryTimeout : 1,
          NumRetries: clustersConfig.RetryPolicy?.NumRetries ? clustersConfig.RetryPolicy.NumRetries : 1,
          RetryBackoffBaseInterval: clustersConfig.RetryPolicy?.RetryBackoffBaseInterval ? clustersConfig.RetryPolicy.RetryBackoffBaseInterval : 1,
        }
      ),
      clustersConfig?.ConnectionSettings?.tcp?.MaxConnections && (
        obj['TcpMaxConnections'] = clustersConfig.ConnectionSettings.tcp.MaxConnections,
        clustersConfig.ConnectionSettings?.http?.MaxRequestsPerConnection && (
          obj['HttpMaxRequestsPerConnection'] = clustersConfig.ConnectionSettings.http.MaxRequestsPerConnection
        ),
        clustersConfig.ConnectionSettings?.http?.MaxPendingRequests && (
          obj['HttpMaxPendingRequests'] = clustersConfig.ConnectionSettings.http.MaxPendingRequests + clustersConfig.ConnectionSettings.tcp.MaxConnections
        )
      ),
      obj
    ))()
  ), null, {}),

  statsKeyCache = new algo.Cache((key) => (
    ((prefix = 'cluster.' + key + '.upstream_rq_retry') => (
      {
        prefixKey: prefix,
        bakoffKey: prefix + '_backoff_exponential',
        limitKey: prefix + '_limit_exceeded',
        successKey: prefix + '_success',
        pendingStatsKey: 'cluster.' + key + '.upstream_rq_pending_overflow',
      }
    ))()
  ), null, {}),

  calcScaleRatio = (n) => (
    n < 1 ? 0 : (n = Math.pow(2, n - 1), n > 10 ? 10 : n)
  )

) => (

  pipy({
    _overflow: false,
    _timestamp: null,
    _retryCount: null,
    _retryPolicy: null,
    _muxHttpOptions: null,
    _clustersConfig: null,
  })

    .import({
      _outMatch: 'outbound-classifier',
      _outTarget: 'outbound-classifier',
      _upstreamClusterName: 'outbound-classifier',
    })

    //
    // Multiplexer for upstream HTTP
    //
    .pipeline()

    .onStart(
      () => void (
        _timestamp = Date.now(),
        outClustersConfigs?.[_upstreamClusterName] && (
          _clustersConfig = clustersConfigCache.get(outClustersConfigs[_upstreamClusterName])
        )
      )
    )

    .branch(
      () => Boolean(_clustersConfig?.HttpMaxPendingRequests), $ => $
        .muxQueue(() => _upstreamClusterName, () => ({
          maxQueue: _clustersConfig.HttpMaxPendingRequests
        }))
        .to($ => $
          .onStart((_, n) => void (_overflow = (n > 1)))
          .branch(
            () => _overflow, $ => $
              .replaceData()
              .replaceMessage([new Message({ overflow: true }), new StreamEnd]),
            $ => $
              .demuxQueue().to($ => $
                .link('upstream-http-request')
              )
          )
        ),
      $ => $
        .link('upstream-http-request')
    )

    .replaceMessage(
      // Circuit breaking for destinations within the mesh
      msg => (
        (_overflow = Boolean(msg.head?.overflow)) ?
          (metrics.sidecarInsideStats[statsKeyCache.get(_upstreamClusterName).pendingStatsKey]++,
            new Message({ status: 503 }, 'Service Unavailable'))
          :
          (msg?.head?.headers && (
            msg.head.headers['server'] = 'pipy',
            msg.head.headers['x-pipy-upstream-service-time'] = Math.ceil(Date.now() - _timestamp)
          ), msg)
      )
    )

    //
    // upstream request
    //
    .pipeline('upstream-http-request')

    .handleMessageStart(
      () => (
        (_retryPolicy = _clustersConfig?.RetryPolicy) && (
          _retryCount = 0
        ),
        _muxHttpOptions = {},
        (_outMatch?.Protocol === 'grpc') && (
          _muxHttpOptions['version'] = 2
        ),
        _clustersConfig?.HttpMaxPendingRequests && (
          _muxHttpOptions['maxQueue'] = _clustersConfig.HttpMaxPendingRequests
        ),
        _clustersConfig?.HttpMaxRequestsPerConnection && (
          _muxHttpOptions['maxMessages'] = _clustersConfig.HttpMaxRequestsPerConnection
        )
      )
    )

    .replay({ 'delay': () => _retryPolicy?.RetryBackoffBaseInterval ? _retryPolicy.RetryBackoffBaseInterval * calcScaleRatio(_retryCount) / 1000.0 : 0 })
    .to($ => $
      .muxHTTP(() => _outTarget, () => _muxHttpOptions).to($ => $.chain())
      .replaceMessage(
        msg => (msg?.head?.status ?
          ((status = msg.head.status, again = false, statsKeys = statsKeyCache.get(_upstreamClusterName)) => (
            (_retryPolicy && status >= _retryPolicy.lowerbound && status <= _retryPolicy.upperbound) && (
              _retryCount < _retryPolicy.NumRetries ? (
                metrics.sidecarInsideStats[statsKeys.prefixKey]++,
                metrics.sidecarInsideStats[statsKeys.bakoffKey]++,
                _retryCount++,
                again = true
              ) : (
                metrics.sidecarInsideStats[statsKeys.limitKey]++
              )
            ),
            (_retryPolicy && _retryCount > 0 && status >= '200' && status <= '299') && (
              metrics.sidecarInsideStats[statsKeys.successKey]++
            ),
            again ? new StreamEnd('Replay') : msg
          ))() : msg
        )
      )
    )

))()