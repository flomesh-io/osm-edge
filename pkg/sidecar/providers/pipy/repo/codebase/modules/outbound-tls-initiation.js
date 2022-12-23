((
  config = pipy.solve('config.js'),
  {
    clusterCache
  } = pipy.solve('modules/metrics.js'),

  certChain = config?.Certificate?.CertChain,
  privateKey = config?.Certificate?.PrivateKey,
  issuingCA = config?.Certificate?.IssuingCA,

  listIssuingCA = (
    (cas = []) => (
      issuingCA && cas.push(new crypto.Certificate(issuingCA)),
      Object.values(config?.Outbound?.TrafficMatches || {}).map(
        a => a.map(
          o => Object.values(o.DestinationIPRanges || {}).map(
            c => c?.SourceCert?.IssuingCA && (
              cas.push(new crypto.Certificate(c?.SourceCert?.IssuingCA))
            )
          )
        )
      ),
      Object.values(config?.Outbound?.ClustersConfigs || {}).forEach(
        c => c?.SourceCert?.IssuingCA && (
          cas.push(new crypto.Certificate(c?.SourceCert?.IssuingCA))
        )
      ),
      cas
    )
  )(),

  forwardMatches = config?.Forward?.ForwardMatches && Object.fromEntries(
    Object.entries(
      forwardMatches).map(
        ([k, v]) => [
          k, new algo.RoundRobinLoadBalancer(v || {})
        ]
      )
  ),

  forwardEgressGateways = config?.Forward?.EgressGateways && Object.fromEntries(
    Object.entries(
      forwardEgressGateways).map(
        ([k, v]) => [
          k, new algo.RoundRobinLoadBalancer(v?.Endpoints || {})
        ]
      )
  ),
) => (

pipy({
  _metrics: null,
  _clusterName: null,
  _egressType: '',
  _egressEndpoint: null,
})

.import({
  __port: 'outbound-main',
  __cert: 'outbound-main',
  __cluster: 'outbound-main',
  __protocol: 'outbound-main',
  __target: 'outbound-main',
  __egressEnable: 'outbound-main',
  __targetObject: 'outbound-http-load-balancing',
  __muxHttpOptions: 'outbound-http-load-balancing',
})

.pipeline()
.branch(
  () => !__target, (
    $=>$.chain()
  ),

  () => __protocol === 'http', (
    $=>$.muxHTTP(() => __targetObject, () => __muxHttpOptions).to(
      $=>$.link('upstream')
    )
  ),

  (
    $=>$.link('upstream')
  )
)

.pipeline('upstream')
.onStart(
  () => void (
    _clusterName = __cluster?.name,
    _metrics = clusterCache.get(_clusterName),
    _metrics.activeConnectionGauge.increase(),

    !__cert && __cluster?.SourceCert && (
      __cluster.SourceCert.OsmIssued && (
        __cert = {CertChain: certChain, PrivateKey: privateKey}
      ) || (
        __cert = __cluster.SourceCert
      )
    ),

    forwardMatches && ((egw = forwardMatches[__port?.EgressForwardGateway || '*']?.next?.()?.id) => (
      egw && (
        _egressType = forwardEgressGateways?.[egw]?.mode || 'http2tunnel',
        _egressEndpoint = forwardEgressGateways?.[egw]?.next?.()?.id
      )
    ))()
    // , console.log('outbound - TLS/__egressEnable/_egressEndpoint/__cert/__target:', Boolean(certChain), __egressEnable, _egressEndpoint, Boolean(__cert), __target)
  )
)
.onEnd(
  () => void (
    _metrics.activeConnectionGauge.decrease()
  )
)
.handleData(
  data => (
    _metrics.sendBytesTotalCounter.increase(data.size)
  )
)
.branch(
  () => __cert, (
    $=>$
    .connectTLS({
      certificate: () => ({
        cert: new crypto.Certificate(__cert.CertChain),
        key: new crypto.PrivateKey(__cert.PrivateKey),
      }),
      trusted: listIssuingCA,
    }).to($=>$.connect(() => __target))
  ),

  () => certChain && !__egressEnable, (
    $=>$
    .connectTLS({
      certificate:() => ({
        cert: new crypto.Certificate(certChain),
        key: new crypto.PrivateKey(privateKey),
      }),
      trusted: issuingCA ? [new crypto.Certificate(issuingCA)] : [],
    }).to($=>$.connect(() => __target))
  ),

  () => __egressEnable && _egressEndpoint, (
    $=>$
    .branch(
      () => _egressType === 'http2tunnel', (
        $ => $
        .connectHTTPTunnel(
          () => new Message({
            method: 'CONNECT',
            path: __target,
          })
        ).to(
          $ => $.muxHTTP(() => __target, { version: 2 }).to(
            $ => $.connect(() => _egressEndpoint)
          )
        )
      ), ($ => $
        .connectSOCKS(
          () => __target,
        ).to($ => $
          .connect(
            () => _egressEndpoint
          )
        )
      )
    )
  ),

  $=>$.connect(() => __target)
)
.handleData(
  data => (
    _metrics.receiveBytesTotalCounter.increase(data.size)
  )
)

))()