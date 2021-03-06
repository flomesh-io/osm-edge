// version: '2022.07.18'
(
  (config = JSON.decode(pipy.load('pipy.json')),
    metrics = pipy.solve('metrics.js'),
    codeMessage = pipy.solve('codes.js'),
    global
  ) => (

    global = {
      debugLogLevel: (config?.Spec?.SidecarLogLevel === 'debug'),
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
      funcShuffle: null,
      codeMessage: codeMessage
    },

    global.funcShuffle = (arg, out, sort) => (
      arg && (() => (
        sort = a => (a.map(e => e).map(() => a.splice(Math.random() * a.length | 0, 1)[0])),
        global.debugLogLevel && console.log('funcShuffle LB in : ', arg),
        out = Object.fromEntries(sort(sort(Object.entries(arg)))),
        global.debugLogLevel && console.log('funcShuffle LB out : ', out)
      ))(),

      out ? out : {}
    ),

    global.funcHttpServiceRouteRules = json => (
      Object.fromEntries(Object.entries(json).map(
        ([name, rule]) => [
          name,
          Object.entries(rule).map(
            ([path, condition]) => ({
              Path_: path, // for debugLogLevel
              Path: new RegExp(path), // HTTP request path
              Methods: condition.Methods && Object.fromEntries(condition.Methods.map(e => [e, true])),
              Headers_: condition?.Headers, // for debugLogLevel
              Headers: condition.Headers && Object.entries(condition.Headers).map(([k, v]) => [k, new RegExp(v)]),
              AllowedServices: condition.AllowedServices && Object.fromEntries(condition.AllowedServices.map(e => [e, true])),
              TargetClusters: condition.TargetClusters && new algo.RoundRobinLoadBalancer(global.funcShuffle(condition.TargetClusters)) // Loadbalancer for services
            })
          )
        ]
      ))
    ),

    global.inTrafficMatches = config?.Inbound?.TrafficMatches && Object.fromEntries(
      Object.entries(config.Inbound.TrafficMatches).map(
        ([port, match]) => [
          port, // local service port
          ({
            Port: match.Port,
            Protocol: match.Protocol,
            HttpHostPort2Service: match.HttpHostPort2Service,
            SourceIPRanges_: match?.SourceIPRanges, // for debugLogLevel
            SourceIPRanges: match.SourceIPRanges && match.SourceIPRanges.map(e => new Netmask(e)),
            TargetClusters: match.TargetClusters && new algo.RoundRobinLoadBalancer(global.funcShuffle(match.TargetClusters)),
            HttpServiceRouteRules: match.HttpServiceRouteRules && global.funcHttpServiceRouteRules(match.HttpServiceRouteRules),
            ProbeTarget: (match.Protocol === 'http') && (!global.probeTarget || !match.SourceIPRanges) && (global.probeTarget = '127.0.0.1:' + port)
          })
        ]
      )
    ),

    global.inClustersConfigs = config?.Inbound?.ClustersConfigs && Object.fromEntries(
      Object.entries(
        config.Inbound.ClustersConfigs).map(
          ([k, v]) => [
            k, (metrics.funcInitClusterNameMetrics(global.namespace, global.kind, global.name, global.pod, k), new algo.RoundRobinLoadBalancer(global.funcShuffle(v)))
          ]
        )
    ),

    global.outTrafficMatches = config?.Outbound?.TrafficMatches && Object.fromEntries(
      Object.entries(config.Outbound.TrafficMatches).map(
        ([port, match]) => [
          port,
          (
            match?.map(
              (o =>
              ({
                Port: o.Port,
                Protocol: o.Protocol,
                ServiceIdentity: o.ServiceIdentity,
                AllowedEgressTraffic: o.AllowedEgressTraffic,
                HttpHostPort2Service: o.HttpHostPort2Service,
                TargetClusters: o.TargetClusters && new algo.RoundRobinLoadBalancer(global.funcShuffle(o.TargetClusters)),
                DestinationIPRanges: o.DestinationIPRanges && o.DestinationIPRanges.map(e => new Netmask(e)),
                HttpServiceRouteRules: o.HttpServiceRouteRules && global.funcHttpServiceRouteRules(o.HttpServiceRouteRules)
              })
              )
            )
          )
        ]
      )
    ),

    // Loadbalancer for endpoints
    global.outClustersConfigs = config?.Outbound?.ClustersConfigs && Object.fromEntries(
      Object.entries(config.Outbound.ClustersConfigs).map(
        ([k, v]) => [
          k, (metrics.funcInitClusterNameMetrics(global.namespace, global.kind, global.name, global.pod, k), new algo.RoundRobinLoadBalancer(global.funcShuffle(v)))
        ]
      )
    ),

    // Initialize probeScheme, probeTarget, probePath
    config?.Spec?.Probes?.LivenessProbes && config.Spec.Probes.LivenessProbes[0]?.httpGet?.port == 15901 &&
    (global.probeScheme = config.Spec.Probes.LivenessProbes[0].httpGet.scheme) && !global.probeTarget &&
    ((global.probeScheme === 'HTTP' && (global.probeTarget = '127.0.0.1:80')) || (global.probeScheme === 'HTTPS' && (global.probeTarget = '127.0.0.1:443'))) &&
    (global.probePath = '/'),
    // PIPY admin port
    global.prometheusTarget = '127.0.0.1:6060',
    global.allowedEndpoints = config?.AllowedEndpoints,
    global.specEnableEgress = config?.Spec?.Traffic?.EnableEgress,

    global.config = config,
    global.metrics = metrics,

    global
  )
)()