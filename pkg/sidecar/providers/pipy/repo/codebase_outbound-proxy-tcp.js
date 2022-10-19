// version: '2022.10.09'
((
  {
    config,
    metrics,
    debugLogLevel,
    tlsCertChain,
    tlsPrivateKey,
    tlsIssuingCA,
    listIssuingCA
  } = pipy.solve('config.js')) => (

  pipy({})

    .import({
      _egressMode: 'main',
      _egressEndpoint: 'main',
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
        debugLogLevel && console.log('outbound connectTLS - TLS/_egressMode/_egressEndpoint/_outSourceCert', Boolean(tlsCertChain), Boolean(_egressMode), _egressEndpoint, Boolean(_outSourceCert)),
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
    .handleData(
      (data) => (
        metrics.receiveBytesTotalCounter.withLabels(_upstreamClusterName).increase(data.size)
      )
    )

))()