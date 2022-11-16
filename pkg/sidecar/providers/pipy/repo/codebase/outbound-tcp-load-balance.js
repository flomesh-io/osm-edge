((
  {
    specEnableEgress,
    outClustersConfigs,
    addIssuingCA,
  } = pipy.solve('config.js'),
  {
    debug,
    shuffle,
  } = pipy.solve('utils.js'),

  targetClustersCache = new algo.Cache((targetClusters) => (
    targetClusters ? new algo.RoundRobinLoadBalancer(shuffle(targetClusters)) : null
  ), null, {}),

  clustersConfigCache = new algo.Cache((clustersConfig) => (
    ((obj = {}) => (
      obj['Endpoints'] = new algo.RoundRobinLoadBalancer(shuffle(clustersConfig.Endpoints)),
      clustersConfig?.SourceCert?.CertChain && clustersConfig?.SourceCert?.PrivateKey && clustersConfig?.SourceCert?.IssuingCA && (
        obj['SourceCert'] = clustersConfig.SourceCert,
        addIssuingCA(clustersConfig.SourceCert.IssuingCA)
      ),
      clustersConfig?.SourceCert?.OsmIssued && tlsCertChain && tlsPrivateKey && (
        obj['SourceCert'] = { CertChain: tlsCertChain, PrivateKey: tlsPrivateKey }
      ),
      obj
    ))()
  ), null, {})

) => (

  pipy({
    egressTargetMap: {}
  })

    .import({
      _outIP: 'outbound-classifier',
      _outPort: 'outbound-classifier',
      _outMatch: 'outbound-classifier',
      _outTarget: 'outbound-classifier',
      _egressEnable: 'outbound-classifier',
      _outSourceCert: 'outbound-classifier',
      _upstreamClusterName: 'outbound-classifier'
    })

    //
    // Analyze inbound HTTP request headers and match routes
    //
    .pipeline()

    .onStart(
      () => (
        // Layer 4 load balance
        _outTarget = (
          (
            // Allow?
            _outMatch &&
            _outMatch.Protocol !== 'http' && _outMatch.Protocol !== 'grpc'
          ) && ((targetCluster, clusterConfig) => (
            (targetCluster = targetClustersCache.get(_outMatch.TargetClusters)) && (
              _upstreamClusterName = targetCluster.next?.()?.id,
              clusterConfig = clustersConfigCache.get(outClustersConfigs?.[_upstreamClusterName]),

              // Egress mTLS certs
              !_outSourceCert && (_outSourceCert = clusterConfig?.SourceCert),
              // Load balance
              clusterConfig?.Endpoints?.next?.()
            )
          ))()
        ),

        // EGRESS mode
        !Boolean(_outTarget) && (specEnableEgress || _outMatch?.AllowedEgressTraffic) && (_outMatch?.Protocol !== 'http' && _outMatch?.Protocol !== 'grpc') && (
          ((target) => (
            target = _outIP + ':' + _outPort,
            _upstreamClusterName = target,
            !egressTargetMap[target] && (egressTargetMap[target] = new algo.RoundRobinLoadBalancer({
              [target]: 100
            })),
            _outTarget = egressTargetMap[target].next(),
            _egressEnable = true
          ))()
        ),

        debug(log => (
          log('outbound _upstreamClusterName: ', _upstreamClusterName),
          log('outbound _outTarget: ', _outTarget?.id)
        )),
        null
      )
    )
    .branch(
      () => Boolean(_outTarget), $ => $
        .chain(),

      $ => $
        .replaceStreamStart(
          new StreamEnd('ConnectionReset')
        )
    )

))()