((
  {
    addIssuingCA,
    tlsCertChain,
    tlsPrivateKey,
    outTrafficMatches,
  } = pipy.solve('config.js'),
  {
    debug,
  } = pipy.solve('utils.js'),

  portMatchesCache = new algo.Cache((portMatches) => (
    portMatches ? portMatches.map(
      (o => (
        {
          Port: o.Port,
          Protocol: o.Protocol,
          AllowedEgressTraffic: o.AllowedEgressTraffic,
          EgressForwardGateway: o.EgressForwardGateway,
          DestinationIPRanges: o.DestinationIPRanges && Object.entries(o.DestinationIPRanges).map(
            ([k, v]) => (
              v?.SourceCert?.IssuingCA && (
                addIssuingCA(v.SourceCert.IssuingCA)
              ),
              {
                netmask: new Netmask(k),
                cert: v?.SourceCert?.OsmIssued && tlsCertChain && tlsPrivateKey ?
                  ({ CertChain: tlsCertChain, PrivateKey: tlsPrivateKey }) : v?.SourceCert
              }
            )
          ),
          JsonObject: o
        }
      ))
    ) : null
  ), null, {})

) => (

  pipy({
  })

    .import({
      _flow: 'main',
    })

    .export('outbound-classifier', {
      _outIP: null,
      _outPort: null,
      _outMatch: null,
      _outTarget: null,
      _egressEnable: false,
      _outSourceCert: null,
      _upstreamClusterName: null,
      _outClustersBreakers: null,
    })

    .pipeline()

    .replaceStreamStart(
      (evt) => (
        _flow = 'outbound',

        // Upstream service port
        _outPort = (__inbound.destinationPort || 0),

        // Upstream service IP
        _outIP = (__inbound.destinationAddress || '127.0.0.1'),

        outTrafficMatches?.[_outPort] && ((portMatch, match) => (
          (portMatch = portMatchesCache.get(outTrafficMatches[_outPort])) && (
            match = (
              // Strict matching Destination IP address
              portMatch.find?.(o => (o.DestinationIPRanges && o.DestinationIPRanges.find(
                e => (e.netmask?.contains?.(_outIP) ? (_outSourceCert = e.cert, true) : false)
              ))) ||
              // EGRESS mode - does not check the IP
              (_egressEnable = true) && portMatch.find?.(o => (!Boolean(o.DestinationIPRanges) &&
                (o.Protocol == 'http' || o.Protocol == 'https' || (o.Protocol == 'tcp' && o.AllowedEgressTraffic))))
            ),
            _outMatch = match?.JsonObject
          )
        ))(),

        debug(log => (
          log('outbound _outMatch: ', _outMatch),
          log('outbound protocol: ', _outMatch?.Protocol)
        )),

        !_outMatch || (_outMatch?.Protocol !== 'http' && _outMatch?.Protocol !== 'grpc') ? evt : null
      )
    )
    .branch(
      () => (_outMatch?.Protocol === 'http' || _outMatch?.Protocol === 'grpc'), $ => $
        .chain([
          'outbound-demux-http.js',
          'outbound-http-routing.js',
          'metrics-http.js',
          'tracing.js',
          'logging.js',
          'outbound-breaker.js',
          'outbound-mux-http.js',
          'metrics-tcp.js',
          'outbound-proxy-tcp.js'
        ]),

      $ => $
        .chain([
          'outbound-tcp-load-balance.js',
          'metrics-tcp.js',
          'outbound-proxy-tcp.js'
        ])
    )

))()