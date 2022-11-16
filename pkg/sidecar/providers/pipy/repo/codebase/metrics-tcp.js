((
  metrics = pipy.solve('metrics-init.js')
) => (

  pipy({
    clusterName: null
  })

    .import({
      _flow: 'main',
      _localClusterName: 'inbound-classifier',
      _upstreamClusterName: 'outbound-classifier'
    })

    .pipeline()

    .onStart(
      () => (
        (_flow === 'inbound') && (
          clusterName = _localClusterName
        ),
        (_flow === 'outbound') && (
          clusterName = _upstreamClusterName
        ),
        null
      )
    )

    .handleData(
      (data) => (
        metrics.sendBytesTotalCounter.withLabels(clusterName).increase(data.size)
      )
    )

    .chain()

    .handleData(
      (data) => (
        metrics.receiveBytesTotalCounter.withLabels(clusterName).increase(data.size)
      )
    )

))()