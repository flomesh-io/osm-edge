((
  {
    shuffle,
    failover,
  } = pipy.solve('utils.js'),

  retryCounter = new stats.Counter('sidecar_cluster_upstream_rq_retry', ['sidecar_cluster_name']),
  retrySuccessCounter = new stats.Counter('sidecar_cluster_upstream_rq_retry_success', ['sidecar_cluster_name']),
  retryLimitCounter = new stats.Counter('sidecar_cluster_upstream_rq_retry_limit_exceeded', ['sidecar_cluster_name']),
  retryOverflowCounter = new stats.Counter('sidecar_cluster_upstream_rq_retry_overflow', ['sidecar_cluster_name']),
  retryBackoffCounter = new stats.Counter('sidecar_cluster_upstream_rq_retry_backoff_exponential', ['sidecar_cluster_name']),
  retryBackoffLimitCounter = new stats.Counter('sidecar_cluster_upstream_rq_retry_backoff_ratelimited', ['sidecar_cluster_name']),

  makeClusterConfig = (clusterConfig) => (
    clusterConfig && (
      (
        obj = {
          targetBalancer: clusterConfig.Endpoints && new algo.RoundRobinLoadBalancer(
            shuffle(Object.fromEntries(Object.entries(clusterConfig.Endpoints).map(([k, v]) => [k, v.Weight])))
          ),
          failoverBalancer: clusterConfig.Endpoints && failover(Object.fromEntries(Object.entries(clusterConfig.Endpoints).map(([k, v]) => [k, v.Weight]))),
          needRetry: Boolean(clusterConfig.RetryPolicy?.NumRetries),
          numRetries: clusterConfig.RetryPolicy?.NumRetries,
          retryStatusCodes: (clusterConfig.RetryPolicy?.RetryOn || '5xx').split(',').reduce(
            (lut, code) => (
              code.endsWith('xx') ? (
                new Array(100).fill(0).forEach((_, i) => lut[(code.charAt(0)|0)*100+i] = true)
              ) : (
                lut[code|0] = true
              ),
              lut
            ),
            []
          ),
          retryBackoffBaseInterval: clusterConfig.RetryPolicy?.RetryBackoffBaseInterval > 1 ? 1 : clusterConfig.RetryPolicy?.RetryBackoffBaseInterval,
          retryCounter: retryCounter.withLabels(clusterConfig.name),
          retrySuccessCounter: retrySuccessCounter.withLabels(clusterConfig.name),
          retryLimitCounter: retryLimitCounter.withLabels(clusterConfig.name),
          retryOverflowCounter: retryOverflowCounter.withLabels(clusterConfig.name),
          retryBackoffCounter: retryBackoffCounter.withLabels(clusterConfig.name),
          retryBackoffLimitCounter: retryBackoffLimitCounter.withLabels(clusterConfig.name),
          muxHttpOptions: {
            version: () => __isHTTP2 ? 2 : 1,
            maxMessages: clusterConfig.ConnectionSettings?.http?.MaxRequestsPerConnection
          },
        },
      ) => (
        obj.retryCounter.zero(),
        obj.retrySuccessCounter.zero(),
        obj.retryLimitCounter.zero(),
        obj.retryOverflowCounter.zero(),
        obj.retryBackoffCounter.zero(),
        obj.retryBackoffLimitCounter.zero(),
        obj
      )
    )()
  ),

  clusterConfigs = new algo.Cache(makeClusterConfig),

  shouldRetry = (statusCode) => (
    _clusterConfig.retryStatusCodes[statusCode] ? (
      (_retryCount < _clusterConfig.numRetries) ? (
        _clusterConfig.retryCounter.increase(),
        _clusterConfig.retryBackoffCounter.increase(),
        _retryCount++,
        true
      ) : (
        _clusterConfig.retryLimitCounter.increase(),
        false
      )
    ) : (
      _retryCount > 0 && _clusterConfig.retrySuccessCounter.increase(),
      false
    )
),

) => pipy({
  _retryCount: 0,
  _clusterConfig: null,
  _failoverObject: null,
})

.import({
  __isHTTP2: 'outbound-main',
  __cluster: 'outbound-main',
  __target: 'outbound-main',
})

.export('outbound-http-load-balancing', {
  __targetObject: null,
  __muxHttpOptions: null,
})

.pipeline()
.onStart(
  () => void (
    (_clusterConfig = clusterConfigs.get(__cluster)) && (
      __targetObject = _clusterConfig.targetBalancer?.next?.(),
      __target = __targetObject?.id,
      __muxHttpOptions = _clusterConfig.muxHttpOptions,
      _clusterConfig.failoverBalancer && (
        _failoverObject = _clusterConfig.failoverBalancer.next()
      )
    )
  )
)

.branch(
  () => _clusterConfig?.needRetry, (
    $=>$
    .replay({
        delay: () => _clusterConfig.retryBackoffBaseInterval * Math.min(10, Math.pow(2, _retryCount-1)|0)
    }).to(
      $=>$
      .chain()
      .replaceMessageStart(
        msg => (
          shouldRetry(msg.head.status) ? new StreamEnd('Replay') : msg
        )
      )
    )
  ),

  () => _failoverObject, (
    $=>$
    .replay({ 'delay': 0 }).to(
      $=>$
      .chain()
      .replaceMessage(
        msg => (
          (
            status = msg?.head?.status
          ) => (
            _failoverObject && (!status || status < '200' || status > '399') ? (
              __targetObject = _failoverObject,
              __target = __targetObject.id,
              _failoverObject = null,
              new StreamEnd('Replay')
            ) : msg
          )
        )()
      )
    )
  ),

  (
    $=>$.chain()
  )
)

)()
