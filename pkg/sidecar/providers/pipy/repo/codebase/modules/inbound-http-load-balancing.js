((

  targetBalancers = new algo.Cache(cluster => new algo.RoundRobinLoadBalancer(cluster?.Endpoints || {})),

) => pipy()

.import({
  __cluster: 'inbound-main',
  __target: 'inbound-main',
})

.pipeline()
.handleStreamStart(
  () => __target = targetBalancers.get(__cluster)?.next?.()
)
.chain()

)()