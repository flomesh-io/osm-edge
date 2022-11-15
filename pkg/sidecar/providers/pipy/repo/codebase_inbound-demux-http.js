(() => (

  pipy({
    _inSessionControl: null
  })

    .pipeline()

    .onStart(
      () => (
        _inSessionControl = {},
        null
      )
    )

    .demuxHTTP().to($ => $
      .handleMessageStart(
        () => (
          _inSessionControl.close = false
        )
      )
      .chain()
      .handleStreamEnd(
        evt => (
          (evt.error === 'ConnectionRefused') && (_inSessionControl.close = true)
        )
      )
    )

    .replaceMessageStart(
      msg => (
        _inSessionControl.close ? new StreamEnd : msg
      )
    )

))()