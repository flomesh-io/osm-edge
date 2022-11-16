((
  {
    inClustersConfigs,
  } = pipy.solve('config.js'),
  {
    debug,
    shuffle,
  } = pipy.solve('utils.js'),

  targetClustersCache = new algo.Cache((targetClusters) => (
    targetClusters ? new algo.RoundRobinLoadBalancer(shuffle(targetClusters)) : null
  ), null, {}),

  clustersConfigCache = new algo.Cache((clustersConfig) => (
    clustersConfig ? new algo.RoundRobinLoadBalancer(shuffle(clustersConfig)) : null
  ), null, {})
  
) => (

  pipy({
  })

    .import({
      _inMatch: 'inbound-classifier',
      _inTarget: 'inbound-classifier',
      _localClusterName: 'inbound-classifier'
    })

    //
    // Analyze inbound HTTP request headers and match routes
    //
    .pipeline()

    .onStart(
      () => (

        // Layer 4 load balance
        _inTarget = (
          (
            // Allow?
            _inMatch &&
            _inMatch.Protocol !== 'http' && _inMatch.Protocol !== 'grpc'
          ) && ((targetCluster) => (
            // Load balance for L4
            (targetCluster = targetClustersCache.get(_inMatch.TargetClusters)) && (
              _localClusterName = targetCluster?.next?.()?.id,
              clustersConfigCache.get(inClustersConfigs?.[_localClusterName])?.next?.()
            )
          ))()
        ),

        debug(log => log('inbound _inTarget: ', _inTarget?.id)),
        null
      )
    )

    .branch(
      () => Boolean(_inTarget), $ => $
        .chain(),

      $ => $
        .replaceStreamStart(
          new StreamEnd('ConnectionReset')
        )
    )

))()