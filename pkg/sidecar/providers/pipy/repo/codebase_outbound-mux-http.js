// version: '2022.09.21'
((
  {
    metrics,
    outClustersConfigs
  } = pipy.solve('config.js')) => (

  pipy({
    _overflow: false,
    _retryCount: null,
    _retryPolicy: null,
    _muxHttpOptions: {}
  })

    .import({
      _outMatch: 'main',
      _outTarget: 'main',
      _upstreamClusterName: 'main'
    })

    //
    // Multiplexer for upstream HTTP
    //
    .pipeline()
    .branch(
      () => Boolean(outClustersConfigs?.[_upstreamClusterName]?.HttpMaxPendingRequests), $ => $
        .muxQueue(() => _upstreamClusterName, () => ({
          maxQueue: outClustersConfigs[_upstreamClusterName].HttpMaxPendingRequests
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
          (metrics.sidecarInsideStats[outClustersConfigs[_upstreamClusterName].HttpMaxPendingStatsKey]++,
            new Message({ status: 503 }, 'Service Unavailable'))
          : msg
      )
    )
    .chain()

    //
    // upstream request
    //
    .pipeline('upstream-http-request')
    .handleMessage(
      () => (
        (_retryPolicy = outClustersConfigs?.[_upstreamClusterName]?.RetryPolicy) && (
          _retryCount = 0
        ),
        (_outMatch?.Protocol === 'grpc') && (
          _muxHttpOptions['version'] = 2
        ),
        outClustersConfigs?.[_upstreamClusterName]?.HttpMaxPendingRequests && (
          _muxHttpOptions['maxQueue'] = outClustersConfigs[_upstreamClusterName].HttpMaxPendingRequests
        ),
        outClustersConfigs?.[_upstreamClusterName]?.HttpMaxRequestsPerConnection && (
          _muxHttpOptions['maxMessages'] = outClustersConfigs[_upstreamClusterName].HttpMaxRequestsPerConnection
        )
      )
    )
    .replay().to($ => $
      .muxHTTP(() => _outTarget, () => _muxHttpOptions).to($ => $.chain(['outbound-proxy-tcp.js']))
      .replaceMessage(
        msg => ((status = msg.head.status, again = false) => (
          (_retryPolicy && status >= _retryPolicy.lowerbound && status <= _retryPolicy.upperbound) && (
            _retryCount < _retryPolicy.NumRetries ? (
              metrics.sidecarInsideStats[_retryPolicy.StatsKeyPrefix]++,
              metrics.sidecarInsideStats[_retryPolicy.StatsKeyPrefix + '_backoff_exponential']++,
              _retryCount++,
              again = true
            ) : (
              metrics.sidecarInsideStats[_retryPolicy.StatsKeyPrefix + '_limit_exceeded']++
            )
          ),
          (_retryPolicy && _retryCount > 0 && status >= '200' && status <= '299') && (
            metrics.sidecarInsideStats[_retryPolicy.StatsKeyPrefix + '_success']++
          ),
          again ? new StreamEnd('Replay') : msg
        ))()
      )
    )

))()
