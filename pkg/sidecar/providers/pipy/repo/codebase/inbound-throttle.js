((
  metrics = pipy.solve('metrics-init.js'),

  initLocalRateLimit = (local) => (
    ((burst, hds = {}) => (
      burst = local.Burst > local.Requests ? local.Burst : local.Requests,
      {
        group: algo.uuid(),
        backlog: local.Backlog > 0 ? local.Backlog : 0,
        quota: new algo.Quota(
          burst, {
          produce: local.Requests,
          per: local.StatTimeWindow > 0 ? local.StatTimeWindow : 1
        }
        ),
        status: local.ResponseStatusCode ? local.ResponseStatusCode : 429,
        headers: (local?.ResponseHeadersToAdd?.forEach?.(h => hds[h.Name] = h.Value), hds)
      }
    ))()
  ),

  initRateLimit = (rateLimit) => (
    rateLimit?.Local?.Requests > 0 ? initLocalRateLimit(rateLimit.Local) : null
  ),

  hostRateLimitCache = new algo.Cache(initRateLimit, null, {}),
  pathRateLimitCache = new algo.Cache(initRateLimit, null, {}),
  headerRateLimitCache = new algo.Cache(initRateLimit, null, {}),

  headerConfigCache = new algo.Cache((rateLimit) => (
    rateLimit ? rateLimit.map(
      o => ({
        Headers: o.Headers && Object.entries(o.Headers).map(([k, v]) => [k, new RegExp(v)]),
        RateLimit: o.RateLimit
      })
    ) : null
  ), null, {})

) => (

  pipy({
    _overflow: false,
    _rateLimit: null,
    _inHostRateLimit: null,
    _inPathRateLimit: null,
    _inHeaderRateLimit: null
  })

    .import({
      _inMatch: 'inbound-classifier',
      _inTarget: 'inbound-classifier',
      _inService: 'inbound-http-routing',
      _inRouteRule: 'inbound-http-routing'
    })

    //
    // Control HTTP throttle.
    //
    .pipeline()

    .handleMessageStart(
      (msg) => (
        // Inbound rate limit quotas.
        _inTarget && ((service = _inMatch.HttpServiceRouteRules?.[_inService], header, rt) => (
          _inHostRateLimit = hostRateLimitCache.get(service?.RateLimit),
          _inPathRateLimit = pathRateLimitCache.get(service?.RouteRules?.[_inRouteRule]?.RateLimit),
          header = headerConfigCache.get(service?.HeaderRateLimits),
          rt = header?.find?.(o => (
            (!o.Headers || o.Headers.every(([k, v]) => v.test(msg?.head?.headers[k] || ''))))),
          rt && (_inHeaderRateLimit = headerRateLimitCache.get(rt?.RateLimit))
        ))()
      )
    )

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
                (_inHostRateLimit.quota.consume(1) != 1) ? [new Message({ overflow: true }), new StreamEnd] : msg
              )
            )
        )
        .handleMessageStart(
          msg => ((_overflow = Boolean(msg.head?.overflow)) && (_rateLimit = _inHostRateLimit))
        ),
      $ => $
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
                    overflow: () => _inPathRateLimit
                  }), new StreamEnd]),
                $ => $
                  .throttleMessageRate(() => _inPathRateLimit.quota)
                  .demuxQueue().to($ => $)
              )
            ),
          $ => $
            .replaceMessage(
              msg => (
                (_inPathRateLimit.quota.consume(1) != 1) ? [new Message({ overflow: true }), new StreamEnd] : msg
              )
            )
        )
        .handleMessageStart(
          msg => ((_overflow = Boolean(msg.head?.overflow)) && (_rateLimit = _inPathRateLimit))
        ),
      $ => $
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
                    overflow: () => _inHeaderRateLimit
                  }), new StreamEnd]),
                $ => $
                  .throttleMessageRate(() => _inHeaderRateLimit.quota)
                  .demuxQueue().to($ => $)
              )
            ),
          $ => $
            .replaceMessage(
              msg => (
                (_inHeaderRateLimit.quota.consume(1) != 1) ? [new Message({ overflow: true }), new StreamEnd] : msg
              )
            )
        )
        .handleMessageStart(
          msg => ((_overflow = Boolean(msg.head?.overflow)) && (_rateLimit = _inHeaderRateLimit))
        ),
      $ => $
    )

    .branch(
      () => _overflow, $ => $
        .replaceMessage(
          () => (
            metrics.sidecarInsideStats['http_local_rate_limiter.http_local_rate_limit.rate_limited'] += 1,
            _rateLimit?.status ?
              new Message({
                status: _rateLimit.status,
                headers: _rateLimit?.headers ? _rateLimit.headers : [],
              }, 'Too Many Requests')
              :
              new Message({
                status: 429
              }, 'Too Many Requests')
          )),
      $ => $
        .chain()
    )

))()