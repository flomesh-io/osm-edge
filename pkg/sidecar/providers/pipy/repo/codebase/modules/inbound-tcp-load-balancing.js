((
  config = pipy.solve('config.js'),

  clusterCache = new algo.Cache(
    (clusterName => (
      (cluster = config?.Inbound?.ClustersConfigs?.[clusterName]) => (
        cluster ? Object.assign({ name: clusterName, Endpoints: cluster }) : null
      )
    )())
  ),

  clusterBalancers = new algo.Cache(cluster => new algo.RoundRobinLoadBalancer(cluster || {})),

  targetBalancers = new algo.Cache(target => new algo.RoundRobinLoadBalancer(target?.Endpoints || {})),

) => pipy()

.import({
  __port: 'inbound',
  __cluster: 'inbound',
  __target: 'inbound',
})

.pipeline()
.handleStreamStart(
  () => (
    (clusterName = clusterBalancers.get(__port?.TargetClusters)?.next?.()?.id) => (
      (__cluster = clusterCache.get(clusterName)) && (
        __target = targetBalancers.get(__cluster)?.next?.()?.id
      )
    )
  )()
)

.chain()

)()