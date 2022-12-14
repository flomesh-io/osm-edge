((
  retryCounter = new stats.Counter('sidecar_cluster_upstream_rq_retry', ['sidecar_cluster_name']),
  retrySuccessCounter = new stats.Counter('sidecar_cluster_upstream_rq_retry_success', ['sidecar_cluster_name']),
  retryLimitCounter = new stats.Counter('sidecar_cluster_upstream_rq_retry_limit_exceeded', ['sidecar_cluster_name']),

  makeClusterConfig = (clusterConfig) => (
    clusterConfig &&
    {
      targetBalancer: new algo.RoundRobinLoadBalancer(
        Object.fromEntries(Object.entries(clusterConfig.Endpoints).map(([k, v]) => [k, v.Weight || 100]))
      ),
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
      retryBackoffBaseInterval: clusterConfig.RetryPolicy?.RetryBackoffBaseInterval || 1, // default 1 second
      retryCounter: retryCounter.withLabels(clusterConfig.clusterName),
      retrySuccessCounter: retrySuccessCounter.withLabels(clusterConfig.clusterName),
      retryLimitCounter: retryLimitCounter.withLabels(clusterConfig.clusterName),
      muxHttpOptions: {
        version: () => __isHTTP2 ? 2 : 1,
        maxMessages: clusterConfig.ConnectionSettings?.http?.MaxRequestsPerConnection
      },
    }
  ),

  clusterConfigs = new algo.Cache(makeClusterConfig),

  shouldRetry = (statusCode) => (
    _clusterConfig.retryStatusCodes[statusCode] ? (
      (_retryCount < _clusterConfig.numRetries) ? (
        _clusterConfig.retryCounter.increase(),
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
  _target: null,
  _retryCount: 0,
  _clusterConfig: null,
})

.import({
  __isHTTP2: 'outbound-main',
  __cluster: 'outbound-main',
  __address: 'outbound-main',
})

.pipeline()
.onStart(
  () => void (
    (_clusterConfig = clusterConfigs.get(__cluster)) && (
      (_target = _clusterConfig.targetBalancer?.next?.()) && (
        __address = _target.id
      )
    )
  )
)
.branch(
  () => !_target, $=>$.chain(),

  () => _clusterConfig.needRetry, (
    $=>$
    .replay({
        delay: () => _clusterConfig.retryBackoffBaseInterval * Math.min(10, Math.pow(2, _retryCount-1)|0)
    }).to(
      $=>$
      .link('upstream')
      .replaceMessageStart(
        msg => (
          shouldRetry(msg.head.status) ? new StreamEnd('Replay') : msg
        )
      )
    )
  ),

  (
    $=>$.link('upstream')
  )
)

.pipeline('upstream')
.muxHTTP(() => _target, () => _clusterConfig.muxHttpOptions).to(
  $=>$.chain()
)

)()