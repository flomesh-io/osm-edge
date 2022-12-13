((
  config = pipy.solve('config.js'),

  clusterBalancers = new algo.Cache(cluster => new algo.RoundRobinLoadBalancer(cluster || {})),

  targetBalancers = new algo.Cache(clusterName =>
    new algo.RoundRobinLoadBalancer(config?.Inbound?.ClustersConfigs?.[clusterName] || {})
  ),

) => pipy({
  _clusterName: null,
  _address: null,
})

.import({
  __port: 'inbound-main',
})

.pipeline()
.handleStreamStart(
  () => (
    (_clusterName = clusterBalancers.get(__port?.TargetClusters)?.next?.()?.id) && (
      _address = targetBalancers.get(_clusterName)?.next?.()?.id
    )
  )
)
.branch(
  () => _address, (
    $=>$
    .connect(() => _address)
  ), (
    $=>$.chain()
  )
)

)()