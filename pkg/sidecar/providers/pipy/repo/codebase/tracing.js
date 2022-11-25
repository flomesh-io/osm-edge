((
  {
    namespace,
    kind,
    name,
    pod,
  } = pipy.solve('config.js'),
  tracing = pipy.solve('tracing-init.js')
) => (

  pipy({
    type: null,
    shared: null,
    target: null,
    protocol: null,
    zipkinData: null,
    clusterName: null,
    httpBytesStruct: null,
  })

    .import({
      _flow: 'main',
      _inTarget: 'inbound-classifier',
      _localClusterName: 'inbound-classifier',
      _outMatch: 'outbound-classifier',
      _outTarget: 'outbound-classifier',
      _upstreamClusterName: 'outbound-classifier'
    })

    .pipeline()

    .onStart(
      () => (
        tracing.logZipkin && void (
          (_flow === 'inbound') && (
            type = 'SERVER',
            shared = true,
            target = _inTarget?.id,
            clusterName = _localClusterName
          ),
          (_flow === 'outbound') && (
            type = 'CLIENT',
            shared = false,
            target = _outTarget.id,
            protocol = _outMatch?.Protocol,
            clusterName = _upstreamClusterName
          ),
          httpBytesStruct = {},
          httpBytesStruct.requestSize = httpBytesStruct.responseSize = 0
        )
      )
    )

    .handleMessage(
      (msg) => (
        tracing.logZipkin && (
          httpBytesStruct.requestSize += msg?.body?.size,
          !shared && tracing.initTracingHeaders(namespace, kind, name, pod, msg.head.headers, protocol),
          zipkinData = tracing.makeZipKinData(name, msg, msg.head.headers, clusterName, type, shared)
        )
      )
    )

    .chain()

    .handleMessage(
      (msg) => (
        tracing.logZipkin && msg?.head?.status && (
          httpBytesStruct.responseSize += msg?.body?.size,
          tracing.save(zipkinData, msg?.head, httpBytesStruct, target)
        )
      )
    )

))()