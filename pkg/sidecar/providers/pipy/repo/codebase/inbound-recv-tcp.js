// version: '2022.08.12'
(() => (
  pipy({
  })

    .import({
      _inMatch: 'main',
      _inTarget: 'main',
      _inBytesStruct: 'main',
      _inSessionControl: 'main'
    })

    //
    // check inbound protocol
    //
    .pipeline()
    .branch(
      () => _inMatch?.Protocol === 'http' || _inMatch?.Protocol === 'grpc', $ => $
        .demuxHTTP().to($ => $
          .handleData(
            (data) => (
              _inBytesStruct.requestSize += data.size
            )
          )
          .replaceMessageStart(
            evt => _inSessionControl.close ? new StreamEnd : evt
          ).chain(['inbound-recv-http.js'])
          .handleData(
            (data) => (
              _inBytesStruct.responseSize += data.size
            )
          )
          .use(['gather.js'], 'after-local-http')
        ),
      () => Boolean(_inTarget), $ => $
        .chain(['inbound-proxy-tcp.js']),
      $ => $
        .replaceStreamStart(
          new StreamEnd('ConnectionReset')
        )
    )

))()