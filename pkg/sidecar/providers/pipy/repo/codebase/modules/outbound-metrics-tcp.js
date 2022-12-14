((
  {
    activeConnectionGauge,
    sendBytesTotalCounter,
    receiveBytesTotalCounter
  } = pipy.solve('modules/metrics.js'),
) => pipy({
  _clusterName: null,
})

.import({
  __cluster: 'outbound-main',
})

.pipeline()
.onStart(
  () => void (
    _clusterName = __cluster?.clusterName,
    activeConnectionGauge.withLabels(_clusterName).increase()
  )
)
.onEnd(
  () => void (
    activeConnectionGauge.withLabels(_clusterName).decrease()
  )
)
.handleData(
  (data) => (
    sendBytesTotalCounter.withLabels(_clusterName).increase(data.size)
  )
)
.chain()
.handleData(
  (data) => (
    receiveBytesTotalCounter.withLabels(_clusterName).increase(data.size)
  )
)

)()