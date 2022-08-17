// version: '2022.08.12'
((
  {
    metrics
  } = pipy.solve('config.js')) => (

  pipy({
  })

    .import({
      _inTarget: 'main',
      _localClusterName: 'main'
    })

    //
    // Connect to local service
    //
    .pipeline()
    .onStart(
      () => (
        metrics.activeConnectionGauge.withLabels(_localClusterName).increase()
      )
    )
    .onEnd(
      () => (
        metrics.activeConnectionGauge.withLabels(_localClusterName).decrease()
      )
    )
    .handleData(
      (data) => (
        metrics.sendBytesTotalCounter.withLabels(_localClusterName).increase(data.size)
      )
    )
    .connect(
      () => _inTarget?.id
    )
    .handleData(
      (data) => (
        metrics.receiveBytesTotalCounter.withLabels(_localClusterName).increase(data.size)
      )
    )

))()