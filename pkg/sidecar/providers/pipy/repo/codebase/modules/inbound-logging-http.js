((
  {
    loggingEnabled,
    makeLoggingData,
    saveLoggingData,
  } = pipy.solve('logging.js')
) => (

pipy({
  loggingData: null
})

.import({
  __target: 'inbound',
  __isIngress: 'inbound',
  __service: 'inbound-http',
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
      saveLoggingData(loggingData, msg, __service?.name, __target, __isIngress, false, 'inbound')
    )
  )
)

))()
