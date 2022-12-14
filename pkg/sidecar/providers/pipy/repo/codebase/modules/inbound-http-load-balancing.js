((

  targetBalancers = new algo.Cache(cluster => new algo.RoundRobinLoadBalancer(cluster?.Endpoints || {})),

) => pipy({
  _target: null,
})

.import({
  __cluster: 'inbound-main',
  __address: 'inbound-main',
})

.pipeline()
.branch(
  () => (_target = targetBalancers.get(__cluster)?.next?.()) && (__address = _target.id), (
    $=>$
    .muxHTTP(() => _target).to(
      $=>$.chain()
    )
  ), (
    $=>$.chain()
  )
)

)()