// version: '2022.09.20'
((
  {
    metrics,
    outClustersConfigs
  } = pipy.solve('config.js')) => (

  pipy({
    _retryCount: null,
    _retryPolicy: null
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
    .handleMessage(
      () => (
        (_retryPolicy = outClustersConfigs?.[_upstreamClusterName]?.RetryPolicy) && (
          _retryCount = 0
        )
      )
    )
    .replay().to($ => $
      .branch(
        () => _outMatch?.Protocol === 'grpc', $ => $
          .muxHTTP(() => _outTarget, {
            version: 2
          }).to($ => $.chain(['outbound-proxy-tcp.js'])),
        $ => $
          .muxHTTP(() => _outTarget).to($ => $.chain(['outbound-proxy-tcp.js']))
      )
      .replaceMessage(
        msg => ((status = msg.head.status, again = false) => (
          (_retryPolicy && status >= _retryPolicy.lowerbound && status <= _retryPolicy.upperbound) && (
            _retryCount < _retryPolicy.NumRetries ? (
              metrics.sidecarInsideStats[_retryPolicy.StatsKeyPrefix] += 1,
              metrics.sidecarInsideStats[_retryPolicy.StatsKeyPrefix + '_backoff_exponential'] += 1,
              _retryCount++,
              again = true
            ) : (
              metrics.sidecarInsideStats[_retryPolicy.StatsKeyPrefix + '_limit_exceeded'] += 1
            )
          ),
          (_retryPolicy && _retryCount > 0 && status >= '200' && status <= '299') && (
            metrics.sidecarInsideStats[_retryPolicy.StatsKeyPrefix + '_success'] += 1
          ),
          again ? new StreamEnd('Replay') : msg
        ))()
      )
    )

    .chain()

))()
