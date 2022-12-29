((
  config = pipy.solve('config.js'),
  specEnableEgress = config?.Spec?.Traffic?.EnableEgress,

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

) => pipy({
  _egressTargetMap: {},
})

.import({
  __port: 'outbound',
  __cluster: 'outbound',
  __target: 'outbound',
  __isEgress: 'outbound',
  __plugins: 'outbound',
})

.pipeline()
.handleStreamStart(
  () => (
    __plugins = __port?.TcpServiceRouteRules?.Plugins,
    ((clusterName = clusterBalancers.get(__port?.TcpServiceRouteRules?.TargetClusters)?.next?.()?.id) => (
      (__cluster = clusterCache.get(clusterName)) && (
        __target = targetBalancers.get(__cluster)?.next?.()?.id
      )
    ))(),
    !__target && (specEnableEgress || __port?.TcpServiceRouteRules?.AllowedEgressTraffic) && (
      (
        target = __inbound.destinationAddress + ':' + __inbound.destinationPort
      ) => (
        __isEgress = true,
        !_egressTargetMap[target] && (_egressTargetMap[target] = new algo.RoundRobinLoadBalancer({
          [target]: 100
        })),
        __target = _egressTargetMap[target].next().id,
        !__cluster && (__cluster = {name: target})
      )
    )()
  )
)

.chain()

)()