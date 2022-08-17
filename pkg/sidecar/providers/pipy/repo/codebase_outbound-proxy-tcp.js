// version: '2022.08.12'
((
  {
    config,
    metrics,
    debugLogLevel,
    tlsCertChain,
    tlsPrivateKey,
    tlsIssuingCA
  } = pipy.solve('config.js')) => (

  pipy({})

    .import({
      _egressMode: 'main',
      _outTarget: 'main',
      _upstreamClusterName: 'main'
    })

    //
    // Connect to upstream service
    //
    .pipeline()
    .onStart(
      () => (
        debugLogLevel && console.log('outbound connectTLS - TLS/_egressMode', Boolean(tlsCertChain), Boolean(_egressMode)),
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
      $ => $
        .connect(() => _outTarget?.id)
    )
    .handleData(
      (data) => (
        metrics.receiveBytesTotalCounter.withLabels(_upstreamClusterName).increase(data.size)
      )
    )

))()