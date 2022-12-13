((
  config = pipy.solve('config.js'),
  certChain = config?.Certificate?.CertChain,
  privateKey = config?.Certificate?.PrivateKey,
  issuingCA = config?.Certificate?.IssuingCA,

  makePortHandler = (port) => (
    (
      destinations = (config?.Outbound?.TrafficMatches[port] || []).map(
        config => ({
          ranges: config.DestinationIPRanges && Object.entries(config.DestinationIPRanges).map(
            ([k, config]) => ({
              mask: new Netmask(k),
              cert: config?.SourceCert?.OsmIssued && certChain && privateKey ? ({
                CertChain: certChain,
                PrivateKey: privateKey,
                IssuingCA: issuingCA,
              }) : config?.SourceCert,
              config
            })
          ),
          config,
        })
      ),

      destinationHandlers = new algo.Cache(
        (address) => (
          (
            cert = null,
            egressEnable = false,
            dst = destinations.find(dst => dst.ranges && dst.ranges.find(r => r.mask.contains(address) && (cert = r.cert, true))) || (
              destinations.find(dst => !dst.ranges && (dst.Protocol !== 'tcp' || dst.AllowedEgressTraffic) && (egressEnable = true))
            ),
            protocol = dst?.config?.Protocol === 'http' || dst?.config?.Protocol === 'grpc' ? 'http' : 'tcp',
            isHTTP2 = dst?.config?.Protocol === 'grpc',
          ) => (
            () => (
              __port = dst?.config,
              __protocol = protocol,
              __isHTTP2 = isHTTP2,
              __cert = cert,
              __egressEnable = egressEnable
            )
          )
        )()
      ),

    ) => (
      () => (
        destinationHandlers.get(__inbound.destinationAddress || '127.0.0.1')()
      )
    )
  )(),

  portHandlers = new algo.Cache(makePortHandler),

) => pipy()

.export('outbound-main', {
  __port: null,
  __protocol: null,
  __isHTTP2: false,
  __cert: null,
  __address: null,
  __clusterName: null,
  __egressEnable: false,
})

.pipeline()
.onStart(
  () => void portHandlers.get(__inbound.destinationPort)()
)
.branch(
  () => __protocol === 'http', (
    $=>$
    .replaceStreamStart()
    .chain([
      'modules/outbound-http-routing.js',
      'modules/outbound-metrics-http.js',
      'modules/outbound-circuit-breaker.js',
      'modules/outbound-http-load-balancing.js',
      'modules/outbound-tls-initiation.js',
      'modules/outbound-http-default.js',
    ])
  ),

  () => __protocol === 'tcp', (
    $=>$.chain([
      'modules/outbound-tcp-load-balancing.js',
      'modules/outbound-tls-initiation.js',
      'modules/outbound-tcp-default.js',
    ])

  ), (
    $=>$.replaceStreamStart(
      new StreamEnd('ConnectionReset')
    )
  )
)

)()
