((
  { initRateLimit } = pipy.solve('utils.js'),

  rateLimitedCounter = new stats.Counter('http_local_rate_limit_route_rate_limited'),

  rateLimitCache = new algo.Cache(initRateLimit),
) => (

pipy({
  _overflow: null,
  _rateLimit: null,
})

.import({
  __route: 'inbound-http-routing',
})

.pipeline()
.branch(
  () => _rateLimit = rateLimitCache.get(__route?.RateLimit), (
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