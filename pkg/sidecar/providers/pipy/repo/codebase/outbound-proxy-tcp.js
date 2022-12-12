// version: '2022.12.11'
((
  {
    config,
    metrics,
    debugLogLevel,
    tlsCertChain,
    tlsPrivateKey,
    tlsIssuingCA,
    listIssuingCA,
    forwardMatches,
    forwardEgressGateways,
  } = pipy.solve('config.js')) => (

  pipy({
    _egressType: '',
    _egressEndpoint: null,
  })

    .import({
      _outMatch: 'main',
      _egressMode: 'main',
      _outSourceCert: 'main',
      _outTarget: 'main',
      _upstreamClusterName: 'main'
    })

    //
    // Connect to upstream service
    //
    .pipeline()
    .onStart(
      () => (
        // Find egress nat gateway
        forwardMatches && ((policy, egw) => (
          policy = _outMatch?.EgressForwardGateway || '*',
          (egw = forwardMatches[policy]?.next?.()?.id) && (
            _egressType = forwardEgressGateways?.[egw]?.mode || 'http2tunnel',
            _egressEndpoint = forwardEgressGateways?.[egw]?.balancer?.next?.()?.id
          )
        ))(),
        debugLogLevel && console.log('outbound connectTLS - TLS/_egressMode/_egressEndpoint/_egressType/_outSourceCert',
          Boolean(tlsCertChain), Boolean(_egressMode), _egressEndpoint, _egressType, Boolean(_outSourceCert)),
        metrics.activeConnectionGauge.withLabels(_upstreamClusterName).increase(),
        config?.outClustersBreakers?.[_upstreamClusterName]?.incConnections?.(),
        null
      )
    )
    .onEnd(
      () => (
        metrics.activeConnectionGauge.withLabels(_upstreamClusterName).decrease(),
        config?.outClustersBreakers?.[_upstreamClusterName]?.decConnections?.()
      )
    )
    .handleMessageStart(
      (msg) => (
        msg?.head?._upstreamClusterName && config?.outClustersBreakers?.[msg?.head?._upstreamClusterName]?.increase?.()
      )
    )
    .handleData(
      (data) => (
        metrics.sendBytesTotalCounter.withLabels(_upstreamClusterName).increase(data.size)
      )
    )
    .branch(
      () => Boolean(_outSourceCert), $ => $
        .connectTLS({
          certificate: () => ({
            cert: new crypto.Certificate(_outSourceCert.CertChain),
            key: new crypto.PrivateKey(_outSourceCert.PrivateKey),
          }),
          trusted: listIssuingCA
        }).to($ => $
          .connect(() => _outTarget?.id)
        ),
      () => (Boolean(tlsCertChain) && !Boolean(_egressMode)), $ => $
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
      () => (Boolean(_egressMode) && Boolean(_egressEndpoint)), $ => $
        .branch(
          () => _egressType === 'http2tunnel', (
          $ => $
            .connectHTTPTunnel(
              () => new Message({
                method: 'CONNECT',
                path: _outTarget?.id,
              })
            ).to(
              $ => $.muxHTTP(() => _outTarget?.id, { version: 2 }).to(
                $ => $.connect(() => _egressEndpoint)
              )
            )
        ), ($ => $
          .connectSOCKS(
            () => _outTarget?.id,
          ).to($ => $
            .connect(
              () => _egressEndpoint
            )
          )
        )
        ),
      $ => $.connect(() => _outTarget?.id)
    )
    .handleData(
      (data) => (
        metrics.receiveBytesTotalCounter.withLabels(_upstreamClusterName).increase(data.size)
      )
    )

))()