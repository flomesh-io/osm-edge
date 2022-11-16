pipy({})

  .import({
    _inMatch: 'inbound-classifier',
    _inTarget: 'inbound-classifier'
  })

  //
  // Multiplexer for local HTTP
  //
  .pipeline()

  .branch(
    () => Boolean(_inTarget) && _inMatch?.Protocol === 'grpc', $ => $
      .muxHTTP(() => _inTarget, {
        version: 2
      }).to($ => $
        .chain()),

    () => Boolean(_inTarget), $ => $
      .muxHTTP(() => _inTarget).to($ => $
        .chain()),

    $ => $
      .replaceMessage(
        new Message({
          status: 403
        }, 'Access denied')
      )
  )