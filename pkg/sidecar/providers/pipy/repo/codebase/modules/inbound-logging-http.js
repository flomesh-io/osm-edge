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
  __address: 'inbound-main',
  __service: 'inbound-http-routing',
  __ingressEnable: 'inbound-http-routing',
})

.pipeline()
.handleMessage(
  (msg) => (
    loggingEnabled && (
      loggingData = makeLoggingData(msg)
    )
  )
)
.chain()
.handleMessage(
  msg => (
    loggingEnabled && (
      padLoggingData(loggingData, msg, __service?.serviceName, __address, __ingressEnable, false, 'inbound'),
      saveLogging(loggingData)
    )
  )
)

))()