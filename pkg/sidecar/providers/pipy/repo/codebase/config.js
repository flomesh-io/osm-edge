(
  (
    config = JSON.decode(pipy.load('config.json')),
    global
  ) => (

    global = {
      config: config,
      isDebugEnabled: (config?.Spec?.SidecarLogLevel === 'debug'),
      namespace: (os.env.POD_NAMESPACE || 'default'),
      kind: (os.env.POD_CONTROLLER_KIND || 'Deployment'),
      name: (os.env.SERVICE_ACCOUNT || ''),
      pod: (os.env.POD_NAME || ''),
      tlsCertChain: config?.Certificate?.CertChain,
      tlsPrivateKey: config?.Certificate?.PrivateKey,
      tlsIssuingCA: config?.Certificate?.IssuingCA,
      specEnableEgress: null,
      inTrafficMatches: null,
      inClustersConfigs: null,
      outTrafficMatches: null,
      outClustersConfigs: null,
      allowedEndpoints: null,
      prometheusTarget: null,
      probeScheme: null,
      probeTarget: null,
      probePath: null,
      forwardMatches: null,
      forwardEgressGateways: null,
      mapIssuingCA: {},
      listIssuingCA: [],
      inboundPort: config?.Inbound?.TrafficMatches ? 15003 : 0,
      outboundPort: config?.Outbound || config?.Spec?.Traffic?.EnableEgress ? 15001 : 0,
    },

    global.inTrafficMatches = config?.Inbound?.TrafficMatches,

    global.inClustersConfigs = config?.Inbound?.ClustersConfigs,

    global.outTrafficMatches = config?.Outbound?.TrafficMatches,

    global.outClustersConfigs = config?.Outbound?.ClustersConfigs,

    global.forwardMatches = config?.Forward?.ForwardMatches,

    global.forwardEgressGateways = config?.Forward?.EgressGateways,

    // Initialize probeScheme, probeTarget, probePath
    config?.Inbound?.TrafficMatches && Object.entries(config.Inbound.TrafficMatches).map(
      ([port, match]) => (
        (match.Protocol === 'http') && (!global.probeTarget || !match.SourceIPRanges) && (global.probeTarget = '127.0.0.1:' + port)
      )
    ),
    config?.Spec?.Probes?.LivenessProbes && config.Spec.Probes.LivenessProbes[0]?.httpGet?.port == 15901 &&
    (global.probeScheme = config.Spec.Probes.LivenessProbes[0].httpGet.scheme) && !global.probeTarget &&
    ((global.probeScheme === 'HTTP' && (global.probeTarget = '127.0.0.1:80')) || (global.probeScheme === 'HTTPS' && (global.probeTarget = '127.0.0.1:443'))) &&
    (global.probePath = '/'),

    // PIPY admin port
    global.prometheusTarget = '127.0.0.1:6060',
    global.allowedEndpoints = config?.AllowedEndpoints,
    global.specEnableEgress = config?.Spec?.Traffic?.EnableEgress,

    global.addIssuingCA = ca => (
      (md5 => (
        md5 = '' + algo.hash(ca),
        !global.mapIssuingCA[md5] && (
          global.listIssuingCA.push(new crypto.Certificate(ca)),
          global.mapIssuingCA[md5] = true
        )
      ))()
    ),

    global.tlsIssuingCA && (
      global.addIssuingCA(global.tlsIssuingCA)
    ),

    global
  )
)()