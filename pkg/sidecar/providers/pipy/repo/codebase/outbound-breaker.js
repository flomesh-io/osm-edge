((
  {
    config
  } = pipy.solve('config.js'),
  breaker = pipy.solve('breaker.js'),

  outClustersBreakers = {}
) => (
  
  config?.Outbound?.ClustersConfigs && Object.entries(config.Outbound.ClustersConfigs).map(
    ([k, v]) => (
      v?.ConnectionSettings && (v.ConnectionSettings?.http?.CircuitBreaking?.StatTimeWindow > 0) &&
      (outClustersBreakers[k] = breaker(
        k,
        v.ConnectionSettings?.tcp?.MaxConnections,
        v.ConnectionSettings?.http?.MaxRequestsPerConnection,
        v.ConnectionSettings?.http?.MaxPendingRequests,
        v.ConnectionSettings?.http?.CircuitBreaking?.MinRequestAmount,
        v.ConnectionSettings?.http?.CircuitBreaking?.StatTimeWindow,
        v.ConnectionSettings?.http?.CircuitBreaking?.SlowTimeThreshold,
        v.ConnectionSettings?.http?.CircuitBreaking?.SlowAmountThreshold,
        v.ConnectionSettings?.http?.CircuitBreaking?.SlowRatioThreshold,
        v.ConnectionSettings?.http?.CircuitBreaking?.ErrorAmountThreshold,
        v.ConnectionSettings?.http?.CircuitBreaking?.ErrorRatioThreshold,
        v.ConnectionSettings?.http?.CircuitBreaking?.DegradedTimeWindow,
        v.ConnectionSettings?.http?.CircuitBreaking?.DegradedStatusCode,
        v.ConnectionSettings?.http?.CircuitBreaking?.DegradedResponseContent))
    )
  ),

  pipy({
    requestTime: null
  })

    .import({
      _upstreamClusterName: 'outbound-classifier'
    })

    .export('outbound-breaker', {
      _outClustersBreakers: outClustersBreakers,
    })

    //
    // Update circuit breaker indicators.
    //
    .pipeline()

    .branch(
      () => outClustersBreakers?.[_upstreamClusterName]?.block?.(), $ => $
        .replaceMessage(
          () => outClustersBreakers[_upstreamClusterName].message()
        ),

      $ => $
        .handleMessageStart(
          () => (
            requestTime = Date.now(),
            outClustersBreakers?.[_upstreamClusterName]?.increase?.()
          )
        )
        .chain()
        .handleMessageStart(
          (msg) => (
            outClustersBreakers?.[_upstreamClusterName]?.checkStatusCode?.(msg?.head?.status),
            outClustersBreakers?.[_upstreamClusterName]?.checkSlow?.((Date.now() - requestTime) / 1000)
          )
        )
    )

    //
    // Periodic calculate circuit breaker ratio.
    //
    .task('1s')
    .onStart(
      () => new Message
    )
    .replaceMessage(
      () => (
        outClustersBreakers && Object.entries(outClustersBreakers).map(
          ([k, v]) => (
            v.sample()
          )
        ),
        new StreamEnd
      )
    )

))()