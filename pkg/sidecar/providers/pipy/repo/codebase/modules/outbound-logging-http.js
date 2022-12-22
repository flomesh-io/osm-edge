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
  __target: 'outbound-main',
  __egressEnable: 'outbound-main',
  __service: 'outbound-http-routing',
})

.pipeline()
.handleMessage(
  msg => (
    loggingEnabled && (
      loggingData = makeLoggingData(msg, __inbound.destinationAddress, __inbound.destinationPort, __inbound.remoteAddress, __inbound.remotePort)
    )
  )
)
.chain()
.handleMessage(
  msg => (
    loggingEnabled && (
      saveLoggingData(loggingData, msg, __service?.name, __target, false, __egressEnable, 'outbound')
    )
  )
)

))()