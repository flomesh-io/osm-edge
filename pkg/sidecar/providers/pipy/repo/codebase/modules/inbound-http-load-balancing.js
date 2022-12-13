((
  config = pipy.solve('config.js'),

  targetBalancers = new algo.Cache(clusterName =>
    new algo.RoundRobinLoadBalancer(config?.Inbound?.ClustersConfigs?.[clusterName] || {})
  ),

) => pipy({
  _target: null,
})

.import({
  __clusterName: 'inbound-main',
})

.pipeline()
.branch(
  () => _target = targetBalancers.get(__clusterName)?.next?.(), (
    $=>$
    .muxHTTP(() => _target).to(
      $=>$.connect(() => _target.id)
    )
  ), (
    $=>$.chain()
  )
)

)()