((
  rateLimitedCounter = new stats.Counter('http_local_rate_limit_route_rate_limited'),

  initRateLimit = (rateLimit) => (
    rateLimit?.Local ? (
      {
        backlog: rateLimit.Local.Backlog || 0,
        quota: new algo.Quota(
          rateLimit.Local.Burst || 0,
          {
            produce: rateLimit.Local.Requests || 0,
            per: rateLimit.Local.StatTimeWindow || 0,
          }
        ),
        response: new Message({
          status: rateLimit.Local.ResponseStatusCode || 429,
          headers: Object.fromEntries((rateLimit.Local.ResponseHeadersToAdd || []).map(({ Name, Value }) => [Name, Value])),
        }),
      }
    ) : null
  ),

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