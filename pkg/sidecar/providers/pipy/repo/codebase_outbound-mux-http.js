// version: '2022.08.12'
pipy({})

  .import({
    _outMatch: 'main',
    _outTarget: 'main'
  })

  //
  // Multiplexer for upstream HTTP
  //
  .pipeline()
  .branch(
    () => _outMatch?.Protocol === 'grpc', $ => $
      .muxHTTP(() => _outTarget, {
        version: 2
      }).to($ => $.chain(['outbound-proxy-tcp.js'])),
    $ => $
      .muxHTTP(() => _outTarget).to($ => $.chain(['outbound-proxy-tcp.js']))
  )
  .chain()