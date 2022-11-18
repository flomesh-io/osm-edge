((
  {
    tlsCertChain,
    tlsPrivateKey,
    tlsIssuingCA,
    listIssuingCA,
    forwardMatches,
    forwardEgressGateways,
  } = pipy.solve('config.js'),

  caShallowCopy = {caList: listIssuingCA},

  metrics = pipy.solve('metrics-init.js'),
  
  {
    debug,
  } = pipy.solve('utils.js'),

  _forwardMatches = forwardMatches && Object.fromEntries(
    Object.entries(
      forwardMatches).map(
        ([k, v]) => [
          k, new algo.RoundRobinLoadBalancer(v || {})
        ]
      )
  ),

  _forwardEgressGateways = forwardEgressGateways && Object.fromEntries(
    Object.entries(
      forwardEgressGateways).map(
        ([k, v]) => [
          k, new algo.RoundRobinLoadBalancer(v?.Endpoints || {})
        ]
      )
  )

) => (

  pipy({
    _egressEndpoint: null,
  })

    .import({
      _outMatch: 'outbound-classifier',
      _outTarget: 'outbound-classifier',
      _egressEnable: 'outbound-classifier',
      _outSourceCert: 'outbound-classifier',
      _upstreamClusterName: 'outbound-classifier',
      _outClustersBreakers: 'outbound-breaker',
    })

    //
    // Connect to upstream service
    //
    .pipeline()

    .onStart(
      () => (
        // Find egress nat gateway
        _forwardMatches && ((policy, egw) => (
          policy = _outMatch?.EgressForwardGateway ? _outMatch?.EgressForwardGateway : '*',
          egw = _forwardMatches[policy]?.next?.()?.id,
          egw && (_egressEndpoint = _forwardEgressGateways?.[egw]?.next?.()?.id)
        ))(),

        metrics.activeConnectionGauge.withLabels(_upstreamClusterName).increase(),
        _outClustersBreakers?.[_upstreamClusterName]?.incConnections?.(),

        debug(log => log('outbound connectTLS - TLS/_egressEnable/_egressEndpoint/_outSourceCert: ', Boolean(tlsCertChain), _egressEnable, _egressEndpoint, Boolean(_outSourceCert))),
        null
      )
    )
    
    .onEnd(
      () => (
        metrics.activeConnectionGauge.withLabels(_upstreamClusterName).decrease(),
        _outClustersBreakers?.[_upstreamClusterName]?.decConnections?.()
      )
    )

    .branch(
      () => Boolean(_outSourceCert), $ => $
        .connectTLS({
          certificate: () => ({
            cert: new crypto.Certificate(_outSourceCert.CertChain),
            key: new crypto.PrivateKey(_outSourceCert.PrivateKey),
          }),
          trusted: caShallowCopy.caList
        }).to($ => $
          .connect(() => _outTarget?.id)
        ),

      () => (Boolean(tlsCertChain) && !_egressEnable), $ => $
        .connectTLS({
          certificate: () => ({
            cert: new crypto.Certificate(tlsCertChain),
            key: new crypto.PrivateKey(tlsPrivateKey),
          }),
          trusted: (!tlsIssuingCA && []) || [
            new crypto.Certificate(tlsIssuingCA),
          ]
        }).to($ => $
          .connect(() => _outTarget?.id)
        ),

      () => (_egressEnable && Boolean(_egressEndpoint)), $ => $
        .connectSOCKS(
          () => _outTarget?.id,
        ).to($ => $
          .connect(
            () => _egressEndpoint
          )
        ),

      $ => $
        .connect(() => _outTarget?.id)
    )
    .chain()

))()
