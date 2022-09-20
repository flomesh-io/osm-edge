// version: '2022.09.20'
(() => (

  pipy({
    _overflow: false
  })

    .import(
      {
        _inHostRateLimit: 'inbound-recv-http',
        _inPathRateLimit: 'inbound-recv-http',
        _inHeaderRateLimit: 'inbound-recv-http'
      }
    )

    //
    // Control HTTP throttle.
    //
    .pipeline()

    .branch(
      () => Boolean(_inHostRateLimit), $ => $
        .branch(
          () => _inHostRateLimit.backlog > 0, $ => $
            .muxQueue(() => _inHostRateLimit.group, () => ({
              maxQueue: _inHostRateLimit.backlog
            })).to($ => $
              .onStart((_, n) => void (_overflow = (n > 1)))
              .branch(
                () => _overflow, $ => $
                  .replaceData()
                  .replaceMessage([new Message({
                    overflow: true
                  }), new StreamEnd]),
                $ => $
                  .throttleMessageRate(() => _inHostRateLimit.quota)
                  .demuxQueue().to($ => $)
              )
            ),
          $ => $
            .replaceMessage(
              msg => (
                (_inHostRateLimit.quota.consume(1) != 1) ? [new Message({ overflow: true, ratelimit: _inHostRateLimit }), new StreamEnd] : msg
              )
            )
        ),
      $ => $
    )
    .handleMessageStart(
      msg => _overflow = Boolean(msg.head?.overflow)
    )

    .branch(
      () => !_overflow && Boolean(_inPathRateLimit), $ => $
        .branch(
          () => _inPathRateLimit.backlog > 0, $ => $
            .muxQueue(() => _inPathRateLimit.group, () => ({
              maxQueue: _inPathRateLimit.backlog
            })).to($ => $
              .onStart((_, n) => void (_overflow = (n > 1)))
              .branch(
                () => _overflow, $ => $
                  .replaceData()
                  .replaceMessage([new Message({
                    overflow: true,
                    ratelimit: 2
                  }), new StreamEnd]),
                $ => $
                  .throttleMessageRate(() => _inPathRateLimit.quota)
                  .demuxQueue().to($ => $)
              )
            ),
          $ => $
            .replaceMessage(
              msg => (
                (_inPathRateLimit.quota.consume(1) != 1) ? [new Message({ overflow: true, ratelimit: _inPathRateLimit }), new StreamEnd] : msg
              )
            )
        ),
      $ => $
    )
    .handleMessageStart(
      msg => _overflow = Boolean(msg.head?.overflow)
    )

    .branch(
      () => !_overflow && Boolean(_inHeaderRateLimit), $ => $
        .branch(
          () => _inHeaderRateLimit.backlog > 0, $ => $
            .muxQueue(() => _inHeaderRateLimit.group, () => ({
              maxQueue: _inHeaderRateLimit.backlog
            })).to($ => $
              .onStart((_, n) => void (_overflow = (n > 1)))
              .branch(
                () => _overflow, $ => $
                  .replaceData()
                  .replaceMessage([new Message({
                    overflow: true,
                    ratelimit: 3
                  }), new StreamEnd]),
                $ => $
                  .throttleMessageRate(() => _inHeaderRateLimit.quota)
                  .demuxQueue().to($ => $)
              )
            ),
          $ => $
            .replaceMessage(
              msg => (
                (_inHeaderRateLimit.quota.consume(1) != 1) ? [new Message({ overflow: true, ratelimit: _inHeaderRateLimit }), new StreamEnd] : msg
              )
            )
        ),
      $ => $
    )

))()