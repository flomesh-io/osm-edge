((
  {
    clusterCache,
  } = pipy.solve('modules/metrics.js'),
) => (

pipy({
  _metrics: null,
  _clusterName: null,
})

.import({
  __protocol: 'inbound-main',
  __isHTTP2: 'inbound-main',
  __cluster: 'inbound-main',
  __target: 'inbound-main',
  __targetObject: 'inbound-http-load-balancing',
})

.pipeline()
.branch(
  () => !__target, (
    $=>$.chain()
  ),

  () => __protocol === 'http', (
    $=>$.muxHTTP(() => __targetObject, { version: () => __isHTTP2 ? 2 : 1 }).to(
      $=>$.link('upstream')
    )
  ),

  (
    $=>$.link('upstream')
  )
)

.pipeline('upstream')
.onStart(
  () => void (
    _clusterName = __cluster?.name,
    _metrics = clusterCache.get(_clusterName),
    _metrics.activeConnectionGauge.increase()
  )
)
.onEnd(
  () => void (
    _metrics.activeConnectionGauge.decrease()
  )
)
.handleData(
  data => (
    _metrics.sendBytesTotalCounter.increase(data.size)
  )
)
.connect(() => __target)
.handleData(
  data => (
    _metrics.receiveBytesTotalCounter.increase(data.size)
  )
)

))()