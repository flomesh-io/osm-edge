((
  config = pipy.solve('config.js'),

  clusterCache = new algo.Cache(
    (clusterName => (
      (cluster = config?.Outbound?.ClustersConfigs?.[clusterName]) => (
        cluster ? Object.assign({ name: clusterName }, cluster) : null
      )
    )())
  ),

  clusterBalancers = new algo.Cache(cluster => new algo.RoundRobinLoadBalancer(cluster || {})),

  targetBalancers = new algo.Cache(target => new algo.RoundRobinLoadBalancer(
    Object.fromEntries(Object.entries(target?.Endpoints || {}).map(([k, v]) => [k, v.Weight || 100]))
  )),

) => pipy()

.import({
  __port: 'outbound-main',
  __cluster: 'outbound-main',
  __target: 'outbound-main',
})

.pipeline()
.handleStreamStart(
  () => (
    ((clusterName = clusterBalancers.get(__port?.TargetClusters)?.next?.()?.id) => (
      (__cluster = clusterCache.get(clusterName)) && (
        __target = targetBalancers.get(__cluster)?.next?.()?.id
      )
    ))()
  )
)
.chain()

)()