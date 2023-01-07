((
  { initRateLimit } = pipy.solve('utils.js'),

  rateLimitedCounter = new stats.Counter('http_local_rate_limit_service_rate_limited'),

  rateLimitCache = new algo.Cache(initRateLimit),
) => (

pipy({
  _overflow: null,
  _rateLimit: null,
})

.import({
  __service: 'inbound-http-routing',
})

.pipeline()
.branch(
  () => _rateLimit = rateLimitCache.get(__service?.RateLimit), (
      $=>$
      .branch(
        () => _rateLimit.backlog > 0, (
          $=>$
          .muxQueue(() => _rateLimit, () => ({ maxQueue: _rateLimit.backlog })).to(
            $=>$
            .onStart((_, n) => void (_overflow = (n > 1)))
            .branch(
              () => _overflow, (
                $=>$
                .replaceData()
                .replaceMessage(
                  () => (
                    rateLimitedCounter.increase(),
                    [_rateLimit.response, new StreamEnd]
                  )
                )
              ), (
                $=>$
                .throttleMessageRate(() => _rateLimit.quota)
                .demuxQueue().to($=>$.chain())
              )
            )
          )
        ), (
          $=>$.replaceMessage(
            msg => (
              _rateLimit.quota.consume(1) !== 1 ? (
                rateLimitedCounter.increase(),
                [_rateLimit.response, new StreamEnd]
              ) : msg
            )
          )
        )
      )
    ), (
      $=>$.chain()
    )
  )

))()