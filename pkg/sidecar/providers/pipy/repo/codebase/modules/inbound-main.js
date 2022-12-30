((
  config = pipy.solve('config.js'),
  {
    inboundL7Chains,
    inboundL4Chains,
  } = pipy.solve('plugins.js'),

  makePortHandler = port => (
    (
      portConfig = config?.Inbound?.TrafficMatches?.[port || 0],
      protocol = portConfig?.Protocol && (portConfig?.Protocol === 'http' || portConfig?.Protocol === 'grpc' ? 'http' : 'tcp'),
      isHTTP2 = portConfig?.Protocol === 'grpc',
      allowedEndpointsLocal = portConfig?.AllowedEndpoints,
      allowedEndpointsGlobal = config?.AllowedEndpoints || {},
      allowedEndpoints = new Set(
        allowedEndpointsLocal
          ? Object.keys(allowedEndpointsLocal).filter(k => k in allowedEndpointsGlobal)
          : Object.keys(allowedEndpointsGlobal)
      ),
      connectionQuota = portConfig?.TcpServiceRouteRules?.L4RateLimit?.Local && (
        new algo.Quota(
          portConfig.TcpServiceRouteRules.L4RateLimit.Local?.Burst || portConfig.TcpServiceRouteRules.L4RateLimit.Local?.Connections || 0,
          {
            produce: portConfig.TcpServiceRouteRules.L4RateLimit.Local?.Connections || 0,
            per: portConfig.TcpServiceRouteRules.L4RateLimit.Local?.StatTimeWindow || 0,
          }
        )
      ),
    ) => (
      !portConfig && (
        () => undefined
      ) || connectionQuota && (
        () => void (
          allowedEndpoints.has(__inbound.remoteAddress || '127.0.0.1') && (connectionQuota.consume(1) === 1) && (
            __port = portConfig,
            __protocol = protocol,
            __isHTTP2 = isHTTP2
          )
        )
      ) || (
        () => void (
          allowedEndpoints.has(__inbound.remoteAddress || '127.0.0.1') && (
            __port = portConfig,
            __protocol = protocol,
            __isHTTP2 = isHTTP2
          )
        )
      )
    )
  )(),

  portHandlers = new algo.Cache(makePortHandler),

) => pipy()

.export('inbound', {
  __port: null,
  __protocol: null,
  __isHTTP2: false,
  __cluster: null,
  __target: null,
  __isIngress: false,
  __plugins: null,
})

.pipeline()
.onStart(
  () => void portHandlers.get(__inbound.destinationPort)()
)
.branch(
  () => __protocol === 'http', (
    $=>$
    .replaceStreamStart()
    .chain(inboundL7Chains)
    /*[
      'modules/inbound-tls-termination.js',
      'modules/inbound-http-routing.js',
      'modules/inbound-metrics-http.js',
      'modules/inbound-tracing-http.js',
      'modules/inbound-logging-http.js',
      'modules/inbound-throttle-service.js',
      'modules/inbound-throttle-route.js',
      'modules/inbound-http-load-balancing.js',
      'modules/inbound-upstream.js',
      'modules/inbound-http-default.js',
    ]*/
  ),

  () => __protocol == 'tcp', (
    $=>$.chain(inboundL4Chains)
    /*[
      'modules/inbound-tls-termination.js',
      'modules/inbound-tcp-load-balancing.js',
      'modules/inbound-upstream.js',
      'modules/inbound-tcp-default.js',
    ]*/
  ),

  (
    $=>$.replaceStreamStart(
      new StreamEnd('ConnectionReset')
    )
  )
)

)()
