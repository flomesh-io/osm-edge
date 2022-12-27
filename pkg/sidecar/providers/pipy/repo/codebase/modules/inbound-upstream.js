((
  config = pipy.solve('config.js'),
  isDebugEnabled = config?.Spec?.SidecarLogLevel === 'debug',
  {
    clusterCache,
  } = pipy.solve('metrics.js'),
) => (

pipy({
  _metrics: null,
  _clusterName: null,
})

.import({
  __protocol: 'inbound',
  __isHTTP2: 'inbound',
  __cluster: 'inbound',
  __target: 'inbound',
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
.branch(
  isDebugEnabled, (
    $=>$
    .handleStreamStart(
      () => (
        console.log('inbound - __protocol, __isHTTP2, __target : ', __protocol, __isHTTP2, __target)
      )
    )
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