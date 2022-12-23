((
  {
    loggingEnabled,
    makeLoggingData,
    saveLoggingData,
  } = pipy.solve('modules/logging.js')
) => (

pipy({
  loggingData: null
})

.import({
  __target: 'inbound-main',
  __service: 'inbound-http-routing',
  __ingressEnable: 'inbound-http-routing',
})

.pipeline()
.handleMessage(
  (msg) => (
    loggingEnabled && (
      loggingData = makeLoggingData(msg, __inbound.remoteAddress, __inbound.remotePort, __inbound.destinationAddress, __inbound.destinationPort)
    )
  )
)
.chain()
.handleMessage(
  msg => (
    loggingEnabled && (
      saveLoggingData(loggingData, msg, __service?.name, __target, __ingressEnable, false, 'inbound')
    )
  )
)

))()
