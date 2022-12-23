((

  targetBalancers = new algo.Cache(cluster => new algo.RoundRobinLoadBalancer(cluster?.Endpoints || {})),

) => pipy()

.import({
  __cluster: 'inbound-main',
  __target: 'inbound-main',
})

.export('inbound-http-load-balancing', {
  __targetObject: null,
})

.pipeline()
.handleStreamStart(
  () => (__targetObject = targetBalancers.get(__cluster)?.next?.()) && (__target = __targetObject.id)
)
.chain()

)()