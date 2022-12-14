((
  config = pipy.solve('config.js'),

  clusterCache = new algo.Cache(
    (clusterName => (
      (cluster = config?.Inbound?.ClustersConfigs?.[clusterName]) => (
        cluster ? Object.assign({ clusterName, Endpoints: cluster }) : null
      )
    )())
  ),

  clusterBalancers = new algo.Cache(cluster => new algo.RoundRobinLoadBalancer(cluster || {})),

  targetBalancers = new algo.Cache(target => new algo.RoundRobinLoadBalancer(target?.Endpoints || {})),

) => pipy()

.import({
  __port: 'inbound-main',
  __cluster: 'inbound-main',
  __address: 'inbound-main',
})

.pipeline()
.handleStreamStart(
  () => (
    (clusterName = clusterBalancers.get(__port?.TargetClusters)?.next?.()?.id) => (
      (__cluster = clusterCache.get(clusterName)) && (
        __address = targetBalancers.get(__cluster)?.next?.()?.id
      )
    )
  )()
)
.chain()

)()