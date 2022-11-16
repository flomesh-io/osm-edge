(() => (
  pipy({
  })

    .pipeline()

    .demuxHTTP().to($ => $
      .chain()
    )

))()