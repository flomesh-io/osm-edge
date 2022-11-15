((
  metrics = pipy.solve('metrics-init.js')
) => (

  pipy({
  })

    .import({
      _inTarget: 'inbound-classifier',
      _localClusterName: 'inbound-classifier'
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

    .connect(
      () => _inTarget?.id
    )

))()