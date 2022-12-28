((
  {
    name,
    tracingEnabled,
    makeZipKinData,
    saveTracing,
  } = pipy.solve('tracing.js')
) => (

pipy({
  zipkinData: null,
  httpBytesStruct: null,
})

.import({
  __cluster: 'inbound',
  __target: 'inbound',
})

.pipeline()
.handleMessage(
  (msg) => (
    tracingEnabled && (
      httpBytesStruct = {},
      httpBytesStruct.requestSize = msg?.body?.size,
      zipkinData = makeZipKinData(name, msg, msg.head.headers, __cluster?.name, 'SERVER', true)
    )
  )
)
.chain()
.handleMessage(
  (msg) => (
    tracingEnabled && (
      httpBytesStruct.responseSize = msg?.body?.size,
      saveTracing(zipkinData, msg?.head, httpBytesStruct, __target)
    )
  )
)

))()