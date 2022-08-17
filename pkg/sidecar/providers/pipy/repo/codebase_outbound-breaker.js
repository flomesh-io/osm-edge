// version: '2022.08.12'
((
  {
    config
  } = pipy.solve('config.js')) => (

  pipy({})

    .import({
      _outRequestTime: 'main',
      _upstreamClusterName: 'main'
    })

    //
    // Update circuit breaker indicators.
    //
    .pipeline()
    .handleMessageStart(
      (msg) => (
        config.outClustersBreakers[_upstreamClusterName]?.checkStatusCode?.(msg?.head?.status),
        config.outClustersBreakers[_upstreamClusterName]?.checkSlow?.((Date.now() - _outRequestTime) / 1000)
      )
    )

))()
