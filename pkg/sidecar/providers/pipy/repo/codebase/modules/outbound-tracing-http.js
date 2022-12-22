((
  {
    namespace,
    kind,
    name,
    pod,
    tracingEnabled,
    initTracingHeaders,
    makeZipKinData,
    saveTracing,
  } = pipy.solve('modules/tracing.js')
) => (

pipy({
  zipkinData: null,
  httpBytesStruct: null,
})

.import({
  __protocol: 'outbound-main',
  __cluster: 'outbound-main',
  __target: 'outbound-main',
})

.pipeline()
.handleMessage(
  (msg) => (
    tracingEnabled && (
      httpBytesStruct = {},
      httpBytesStruct.requestSize = msg?.body?.size,
      initTracingHeaders(namespace, kind, name, pod, msg.head.headers, __protocol),
      zipkinData = makeZipKinData(name, msg, msg.head.headers, __cluster?.name, 'CLIENT', false)
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