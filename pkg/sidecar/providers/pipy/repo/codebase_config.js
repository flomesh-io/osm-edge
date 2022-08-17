// version: '2022.08.15'
(
  (config = JSON.decode(pipy.load('pipy.json')),
    metrics = pipy.solve('metrics.js'),
    breaker = pipy.solve('breaker.js'),
    codeMessage = pipy.solve('codes.js'),
    global
  ) => (

    config.outClustersBreakers = {},

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
      out || {}
    ),

    global.funcInboundHttpServiceRouteRules = json => (
      Object.fromEntries(Object.entries(json).map(
        ([name, rule]) => [
          name,
          {
            RouteRules: rule?.RouteRules && Object.entries(rule.RouteRules).map(
              ([path, condition], obj) => (
                obj = {
                  Path_: path, // for debugLogLevel
                  Path: new RegExp(path), // HTTP request path
                  Methods: condition.Methods && Object.fromEntries(condition.Methods.map(e => [e, true])),
                  Headers_: condition?.Headers, // for debugLogLevel
                  Headers: condition.Headers && Object.entries(condition.Headers).map(([k, v]) => [k, new RegExp(v)]),
                  AllowedServices: condition.AllowedServices && Object.fromEntries(condition.AllowedServices.map(e => [e, true])),
                  TargetClusters: condition.TargetClusters && new algo.RoundRobinLoadBalancer(global.funcShuffle(condition.TargetClusters)) // Loadbalancer for services
                },
                obj.RateLimit = (condition?.RateLimit?.Local?.Requests > 0) && ((burst) => (
                  burst = condition?.RateLimit?.Local?.Burst > condition.RateLimit.Local.Requests ? condition.RateLimit.Local.Burst : condition.RateLimit.Local.Requests,
                  {
                    group: algo.uuid(),
                    backlog: condition?.RateLimit?.Local?.Backlog > 0 ? condition.RateLimit.Local.Backlog : 0,
                    quota: new algo.Quota(
                      burst,
                      { produce: condition.RateLimit.Local.Requests, per: condition?.RateLimit?.Local?.StatTimeWindow > 0 ? condition.RateLimit.Local.StatTimeWindow : 1 }
                    )
                  }
                ))(),
                obj
              )
            ),
            RateLimit: (rule?.RateLimit?.Local?.Requests > 0) && ((burst) => (
              burst = rule?.RateLimit?.Local?.Burst > rule.RateLimit.Local.Requests ? rule.RateLimit.Local.Burst : rule.RateLimit.Local.Requests,
              {
                group: algo.uuid(),
                backlog: rule?.RateLimit?.Local?.Backlog > 0 ? rule.RateLimit.Local.Backlog : 0,
                quota: new algo.Quota(
                  burst,
                  { produce: rule.RateLimit.Local.Requests, per: rule?.RateLimit?.Local?.StatTimeWindow > 0 ? rule.RateLimit.Local.StatTimeWindow : 1 }
                )
              }
            ))(),
            HeaderRateLimits: rule?.HeaderRateLimits && rule.HeaderRateLimits.map(
              o => (
                {
                  Headers: o.Headers && Object.entries(o.Headers).map(([k, v]) => [k, new RegExp(v)]),
                  RateLimit: (o?.RateLimit?.Local?.Requests > 0) && ((burst) => (
                    burst = o?.RateLimit?.Local?.Burst > o.RateLimit.Local.Requests ? o.RateLimit.Local.Burst : o.RateLimit.Local.Requests,
                    {
                      group: algo.uuid(),
                      backlog: o?.RateLimit?.Local?.Backlog > 0 ? o.RateLimit.Local.Backlog : 0,
                      quota: new algo.Quota(
                        burst,
                        { produce: o.RateLimit.Local.Requests, per: o?.RateLimit?.Local?.StatTimeWindow > 0 ? o.RateLimit.Local.StatTimeWindow : 1 }
                      )
                    }
                  ))(),
                }
              )
            )
          }
        ]
      ))
    ),

    global.inTrafficMatches = config?.Inbound?.TrafficMatches && Object.fromEntries(
      Object.entries(config.Inbound.TrafficMatches).map(
        ([port, match], obj) => [
          port, // local service port
          (
            obj = {
              Port: match.Port,
              Protocol: match.Protocol,
              HttpHostPort2Service: match.HttpHostPort2Service,
              SourceIPRanges_: match?.SourceIPRanges, // for debugLogLevel
              SourceIPRanges: match.SourceIPRanges && match.SourceIPRanges.map(e => new Netmask(e)),
              TargetClusters: match.TargetClusters && new algo.RoundRobinLoadBalancer(global.funcShuffle(match.TargetClusters)),
              HttpServiceRouteRules: match.HttpServiceRouteRules && global.funcInboundHttpServiceRouteRules(match.HttpServiceRouteRules),
              ProbeTarget: (match.Protocol === 'http') && (!global.probeTarget || !match.SourceIPRanges) && (global.probeTarget = '127.0.0.1:' + port)
            },
            (match?.RateLimit?.Local?.Connections > 0) && (((array) => (
              array = (new Array(match?.RateLimit?.Local?.Connections > 1000 ? 1000 : match?.RateLimit?.Local?.Connections)).fill('connection_').map((e, index) => e + index),
              obj.RateLimit = new algo.LeastWorkLoadBalancer(Object.fromEntries(array.map(k => [k, 1]))),
              obj.RateLimitObject = Object.fromEntries(array.map(k => [k, new String(k)]))
            ))()),
            obj
          )
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

    global.funcOutboundHttpServiceRouteRules = json => (
      Object.fromEntries(Object.entries(json).map(
        ([name, rule]) => [
          name,
          Object.entries(rule).map(
            ([path, condition]) => (
              {
                Path_: path, // for debugLogLevel
                Path: new RegExp(path), // HTTP request path
                Methods: condition.Methods && Object.fromEntries(condition.Methods.map(e => [e, true])),
                Headers_: condition?.Headers, // for debugLogLevel
                Headers: condition.Headers && Object.entries(condition.Headers).map(([k, v]) => [k, new RegExp(v)]),
                AllowedServices: condition.AllowedServices && Object.fromEntries(condition.AllowedServices.map(e => [e, true])),
                TargetClusters: condition.TargetClusters && new algo.RoundRobinLoadBalancer(global.funcShuffle(condition.TargetClusters)) // Loadbalancer for services
              }
            )
          )
        ]
      ))
    ),

    global.outTrafficMatches = config?.Outbound?.TrafficMatches && Object.fromEntries(
      Object.entries(config.Outbound.TrafficMatches).map(
        ([port, match]) => [
          port,
          (
            match?.map(
              (o =>
              (
                {
                  Port: o.Port,
                  Protocol: o.Protocol,
                  ServiceIdentity: o.ServiceIdentity,
                  AllowedEgressTraffic: o.AllowedEgressTraffic,
                  HttpHostPort2Service: o.HttpHostPort2Service,
                  TargetClusters: o.TargetClusters && new algo.RoundRobinLoadBalancer(global.funcShuffle(o.TargetClusters)),
                  DestinationIPRanges: o.DestinationIPRanges && o.DestinationIPRanges.map(e => new Netmask(e)),
                  HttpServiceRouteRules: o.HttpServiceRouteRules && global.funcOutboundHttpServiceRouteRules(o.HttpServiceRouteRules)
                }
              )
              )
            )
          )
        ]
      )
    ),

    // Loadbalancer for endpoints
    global.outClustersConfigs = config?.Outbound?.ClustersConfigs && Object.fromEntries(
      Object.entries(config.Outbound.ClustersConfigs).map(
        ([k, v]) => (
          v?.ConnectionSettings && (v.ConnectionSettings?.http?.CircuitBreaking?.StatTimeWindow > 0) &&
          (config.outClustersBreakers[k] = breaker(
            k,
            v.ConnectionSettings?.tcp?.MaxConnections,
            v.ConnectionSettings?.http?.MaxRequestsPerConnection,
            v.ConnectionSettings?.http?.MaxPendingRequests,
            v.ConnectionSettings?.http?.CircuitBreaking?.StatTimeWindow,
            v.ConnectionSettings?.http?.CircuitBreaking?.SlowTimeThreshold,
            v.ConnectionSettings?.http?.CircuitBreaking?.SlowAmountThreshold,
            v.ConnectionSettings?.http?.CircuitBreaking?.SlowRatioThreshold,
            v.ConnectionSettings?.http?.CircuitBreaking?.ErrorAmountThreshold,
            v.ConnectionSettings?.http?.CircuitBreaking?.ErrorRatioThreshold,
            v.ConnectionSettings?.http?.CircuitBreaking?.DegradedTimeWindow,
            v.ConnectionSettings?.http?.CircuitBreaking?.DegradedStatusCode,
            v.ConnectionSettings?.http?.CircuitBreaking?.DegradedResponseContent)),
          [
            k, ((obj, array) => (
              obj = {},
              v.ConnectionSettings?.tcp?.MaxConnections && (
                array = (new Array(v.ConnectionSettings?.tcp?.MaxConnections > 1000 ? 1000 : v.ConnectionSettings?.tcp?.MaxConnections)).fill('connection_').map((e, index) => e + index),
                obj.RateLimit = new algo.LeastWorkLoadBalancer(Object.fromEntries(array.map(k => [k, 1]))),
                obj.RateLimitObject = Object.fromEntries(array.map(k => [k, new String(k)]))
              ),
              obj.Endpoints = new algo.RoundRobinLoadBalancer(global.funcShuffle(v.Endpoints)),
              metrics.funcInitClusterNameMetrics(global.namespace, global.kind, global.name, global.pod, k),
              obj
            ))()
          ]
        )
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
