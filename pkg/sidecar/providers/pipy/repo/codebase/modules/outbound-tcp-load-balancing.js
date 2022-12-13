((
  config = pipy.solve('config.js'),

  clusterBalancers = new algo.Cache(cluster => new algo.RoundRobinLoadBalancer(cluster || {})),

  targetBalancers = new algo.Cache(clusterName =>
    new algo.RoundRobinLoadBalancer(config?.Outbound?.ClustersConfigs?.[clusterName]?.Endpoints || {})
  ),

) => pipy()

.import({
  __port: 'outbound-main',
  __address: 'outbound-main',
  __clusterName: 'outbound-main',
})

.pipeline()
.handleStreamStart(
  () => (
    (__clusterName = clusterBalancers.get(__port?.TargetClusters)?.next?.()?.id) && (
      __address = targetBalancers.get(__clusterName)?.next?.().id
    )
  )
)
.chain()

)()