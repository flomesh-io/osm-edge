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
  __target: 'outbound',
  __isEgress: 'outbound',
  __service: 'outbound-http',
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
      saveLoggingData(loggingData, msg, __service?.name, __target, false, __isEgress, 'outbound')
    )
  )
)

))()