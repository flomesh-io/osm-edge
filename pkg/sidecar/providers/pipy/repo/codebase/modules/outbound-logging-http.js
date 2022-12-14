((
  {
    loggingEnabled,
    makeLoggingData,
    padLoggingData,
    saveLogging,
  } = pipy.solve('modules/logging.js')
) => (

pipy({
  loggingData: null
})

.import({
  __address: 'outbound-main',
  __egressEnable: 'outbound-main',
  __service: 'outbound-http-routing',
})

.pipeline()
.handleMessage(
  msg => (
    loggingEnabled && (
      loggingData = makeLoggingData(msg)
    )
  )
)
.chain()
.handleMessage(
  msg => (
    loggingEnabled && (
      padLoggingData(loggingData, msg, __service?.serviceName, __address, false, __egressEnable, 'outbound'),
      saveLogging(loggingData)
    )
  )
)

))()