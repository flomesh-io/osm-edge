// version: '2022.05.13'
(config => (
  (
    specEnableEgress,
    inTrafficMatches,
    inClustersConfigs,
    outTrafficMatches,
    outClustersConfigs,
    allowedEndpoints,
    prometheusTarget,
    probeScheme,
    probeTarget,
    probePath,
    funcHttpServiceRouteRules
  ) => (

    funcHttpServiceRouteRules = json => (
      Object.fromEntries(Object.entries(json).map(
        ([name, rule]) => [
          name,
          Object.entries(rule).map(
            ([path, condition]) => ({
              Path: new RegExp(path), // HTTP request path
              Methods: condition.Methods && Object.fromEntries(condition.Methods.map(e => [e, true])),
              Headers: condition.Headers && Object.entries(condition.Headers).map(([k, v]) => [k, new RegExp(v)]),
              AllowedServices: condition.AllowedServices && Object.fromEntries(condition.AllowedServices.map(e => [e, true])),
              TargetClusters: condition.TargetClusters && new algo.RoundRobinLoadBalancer(condition.TargetClusters) // Loadbalancer for services
            })
          )
        ]
      ))
    ),

    inTrafficMatches = config?.Inbound?.TrafficMatches && Object.fromEntries(
      Object.entries(config.Inbound.TrafficMatches).map(
        ([port, match]) => [
          port, // local service port
          ({
            Port: match.Port,
            Protocol: match.Protocol,
            HttpHostPort2Service: match.HttpHostPort2Service,
            SourceIPRanges: match.SourceIPRanges && match.SourceIPRanges.map(e => new Netmask(e)),
            TargetClusters: match.TargetClusters && new algo.RoundRobinLoadBalancer(match.TargetClusters),
            HttpServiceRouteRules: match.HttpServiceRouteRules && funcHttpServiceRouteRules(match.HttpServiceRouteRules),
            ProbeTarget: (match.Protocol === 'http') && (!probeTarget || !match.SourceIPRanges) && (probeTarget = '127.0.0.1:' + port)
          })
        ]
      )
    ),

    inClustersConfigs = config?.Inbound?.ClustersConfigs && Object.fromEntries(
      Object.entries(
        config.Inbound.ClustersConfigs).map(
        ([k, v]) => [
          k, (new algo.RoundRobinLoadBalancer(v))
        ]
      )
    ),

    outTrafficMatches = config?.Outbound?.TrafficMatches && Object.fromEntries(
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
                  TargetClusters: o.TargetClusters && new algo.RoundRobinLoadBalancer(o.TargetClusters),
                  DestinationIPRanges: o.DestinationIPRanges && o.DestinationIPRanges.map(e => new Netmask(e)),
                  HttpServiceRouteRules: o.HttpServiceRouteRules && funcHttpServiceRouteRules(o.HttpServiceRouteRules)
                })
              )
            )
          )
        ]
      )
    ),

    // Loadbalancer for endpoints
    outClustersConfigs = config?.Outbound?.ClustersConfigs && Object.fromEntries(
      Object.entries(config.Outbound.ClustersConfigs).map(
        ([k, v]) => [
          k, (new algo.RoundRobinLoadBalancer(v))
        ]
      )
    ),

    // Initialize probeScheme, probeTarget, probePath
    config?.Spec?.Probes?.LivenessProbes && config.Spec.Probes.LivenessProbes[0]?.httpGet?.port == 15901 &&
    (probeScheme = config.Spec.Probes.LivenessProbes[0].httpGet.scheme) && !probeTarget &&
    ((probeScheme === 'HTTP' && (probeTarget = '127.0.0.1:80')) || (probeScheme === 'HTTPS' && (probeTarget = '127.0.0.1:443'))) &&
    (probePath = '/'),

    specEnableEgress = config?.Spec?.Traffic?.EnableEgress,

    allowedEndpoints = config?.AllowedEndpoints,

    // PIPY admin port
    prometheusTarget = '127.0.0.1:6060',

    pipy({
      _targetCount: new stats.Counter('lbtarget_cnt', ['target']),
      _inMatch: undefined,
      _inTarget: undefined,
      _inSessionControl: null,
      _outIP: undefined,
      _outPort: undefined,
      _outMatch: undefined,
      _outTarget: undefined,
      _outSessionControl: null
    })

    // inbound
    .listen(config?.Inbound?.TrafficMatches ? 15003 : 0, {
      'transparent': true,
      'closeEOF': false
      // 'readTimeout': '5s'
    })
    .handleStreamStart(
      () => (
        // Find a match by destination port
        _inMatch = (
          allowedEndpoints?.[__inbound.remoteAddress] &&
          inTrafficMatches?.[__inbound.destinationPort || 0]
        ),

        // Check client address against the whitelist
        _inMatch?.AllowedEndpoints &&
        _inMatch.AllowedEndpoints[__inbound.remoteAddress] === undefined && (
          _inMatch = null
        ),

        // Layer 4 load balance
        _inTarget = (
          (
            // Allow?
            _inMatch &&
            _inMatch.Protocol !== 'http'
          ) && (
            // Load balance
            inClustersConfigs?.[
              _inMatch.TargetClusters?.select?.()
            ]?.select?.()
          )
        ),

        // Session termination control
        _inSessionControl = {
          close: false
        }
      )
    )
    .link(
      'http_in', () => _inMatch?.Protocol === 'http',
      'connection_in', () => Boolean(_inTarget),
      'deny_in'
    )

    //
    // HTTP proxy for inbound
    //
    .pipeline('http_in')
    .demuxHTTP('inbound')
    .replaceMessageStart(
      evt => _inSessionControl.close ? new StreamEnd : evt
    )

    //
    // Analyze inbound HTTP request headers and match routes
    //
    .pipeline('inbound')
    .handleMessageStart(
      (msg) => (
        ((service, match, headers) => (
          headers = msg.head.headers,

          // Find the service
          service = (
            // When found in SourceIPRanges, service is '*'
            (_inMatch?.SourceIPRanges?.find?.(e => e.contains(__inbound.remoteAddress)) && '*') ||
            // When serviceidentity is present, service is headers.host
            (headers.serviceidentity && _inMatch?.HttpHostPort2Service?.[headers.host])
          ),

          // Find a match by the service's route rules
          match = _inMatch.HttpServiceRouteRules?.[service]?.find(o => (
            // Match methods
            (!o.Methods || o.Methods[msg.head.method]) &&
            // Match service whitelist
            (!o.AllowedServices || o.AllowedServices[headers.serviceidentity]) &&
            // Match path pattern
            o.Path.test(msg.head.path) &&
            // Match headers
            (!o.Headers || o.Headers.every(([k, v]) => v.test(headers[k] || '')))
          )),

          // Layer 7 load balance
          _inTarget = (
            inClustersConfigs[
              match?.TargetClusters?.select?.()
            ]?.select?.()
          ),

          // Close sessions from any HTTP proxies
          !_inTarget && headers['x-forwarded-for'] && (
            _inSessionControl.close = true
          )
        ))()
      )
    )
    .link(
      'request_in', () => Boolean(_inTarget),
      'deny_in_http'
    )

    //
    // Multiplexing access to HTTP service
    //
    .pipeline('request_in')
    .muxHTTP(
      'connection_in', () => _inTarget
    )

    //
    // Connect to local service
    //
    .pipeline('connection_in')
    .connect(
      () => _inTarget
    )

    //
    // Respond to inbound HTTP with 403
    //
    .pipeline('deny_in_http')
    .replaceMessage(
      new Message({
        status: 403
      }, 'Access denied')
    )

    //
    // Close inbound TCP with RST
    //
    .pipeline('deny_in')
    .replaceStreamStart(
      new StreamEnd('ConnectionReset')
    )

    // outbound
    .listen(config?.Outbound || config?.Spec?.Traffic?.EnableEgress ? 15001 : 0, {
      'transparent': true,
      'closeEOF': false
      // 'readTimeout': '5s'
    })
    .handleStreamStart(
      () => (
        // Upstream service port
        _outPort = (__inbound.destinationPort || 0),

        // Upstream service IP
        _outIP = (__inbound.destinationAddress || '127.0.0.1'),

        _outMatch = (outTrafficMatches && outTrafficMatches[_outPort] && (
          // Strict matching Destination IP address
          outTrafficMatches[_outPort].find(o => (o.DestinationIPRanges && o.DestinationIPRanges.find(e => e.contains(_outIP)))) ||
          // EGRESS mode - does not check the IP
          outTrafficMatches[_outPort].find(o => (!Boolean(o.DestinationIPRanges) &&
            (o.Protocol == 'http' || o.Protocol == 'https' || (o.Protocol == 'tcp' && o.AllowedEgressTraffic))))
        )),

        // Layer 4 load balance
        _outTarget = (
          (
            // Allow?
            _outMatch &&
            _outMatch.Protocol !== 'http'
          ) && (
            // Load balance
            outClustersConfigs?.[
              _outMatch.TargetClusters?.select?.()
            ]?.select?.()
          )
        ),

        // EGRESS mode
        !Boolean(_outTarget) && (specEnableEgress || _outMatch?.AllowedEgressTraffic) && (_outMatch?.Protocol !== 'http') && (
          _outTarget = _outIP + ':' + _outPort
        ),

        _outSessionControl = {
          close: false
        }
      )
    )
    .link(
      'http_out', () => _outMatch?.Protocol === 'http',
      'connection_out', () => Boolean(_outTarget),
      'deny_out'
    )

    //
    // HTTP proxy for outbound
    //
    .pipeline('http_out')
    .demuxHTTP('outbound')
    .replaceMessageStart(
      evt => _outSessionControl.close ? new StreamEnd : evt
    )

    //
    // Analyze outbound HTTP request headers and match routes
    //
    .pipeline('outbound')
    .handleMessageStart(
      (msg) => (
        ((service, route, match, headers) => (
          headers = msg.head.headers,

          service = _outMatch.HttpHostPort2Service?.[headers.host],

          // Find route by HTTP host
          route = service && _outMatch.HttpServiceRouteRules?.[service],

          // Find a match by the service's route rules
          match = route?.find(o => (
            // Match methods
            (!o.Methods || o.Methods[msg.head.method]) &&
            // Match service whitelist
            (!o.AllowedServices || o.AllowedServices[headers.serviceidentity]) &&
            // Match path pattern
            o.Path.test(msg.head.path) &&
            // Match headers
            (!o.Headers || o.Headers.every(([k, v]) => v.test(headers[k] || '')))
          )),

          // Layer 7 load balance
          _outTarget = (
            outClustersConfigs[
              match?.TargetClusters?.select?.()
            ]?.select?.()
          ),

          // Add serviceidentity for request authentication
          _outTarget && (headers['serviceidentity'] = _outMatch.ServiceIdentity),

          // EGRESS mode
          !_outTarget && (specEnableEgress || _outMatch?.AllowedEgressTraffic) && (
            _outTarget = _outIP + ':' + _outPort
          ),

          // Loadbalancer metrics
          _outTarget && _targetCount.withLabels(_outTarget).increase()
        ))()
      )
    )
    .link(
      'request_out', () => Boolean(_outTarget),
      'deny_out_http'
    )

    //
    // Multiplexing access to HTTP service
    //
    .pipeline('request_out')
    .muxHTTP(
      'connection_out', () => _outTarget
    )

    //
    // Connect to upstream service
    //
    .pipeline('connection_out')
    .connect(
      () => _outTarget
    )

    //
    // Respond to outbound HTTP with 403
    //
    .pipeline('deny_out_http')
    .replaceMessage(
      new Message({
        status: 403
      }, 'Access denied')
    )

    //
    // Close outbound TCP with RST
    //
    .pipeline('deny_out')
    .replaceStreamStart(
      new StreamEnd('ConnectionReset')
    )

    // liveness probe
    .listen(probeScheme ? 15901 : 0)
    .link(
      'http_liveness', () => probeScheme === 'HTTP',
      'connection_liveness', () => Boolean(probeTarget),
      'deny_liveness'
    )

    //
    // HTTP server for liveness probe
    //
    .pipeline('http_liveness')
    .demuxHTTP('message_liveness')

    //
    // rewrite request URL
    //
    .pipeline('message_liveness')
    .handleMessageStart(
      msg => (
        msg.head.path === '/osm-liveness-probe' && (msg.head.path = '/liveness'),
        probePath && (msg.head.path = probePath)
      )
    )
    .muxHTTP('connection_liveness', probeTarget)

    //
    // connect to the app port
    //
    .pipeline('connection_liveness')
    .connect(() => probeTarget)

    //
    // No target detected, access denied.
    //
    .pipeline('deny_liveness')
    .replaceStreamStart(
      new StreamEnd('ConnectionReset')
    )

    // readiness probe
    .listen(probeScheme ? 15902 : 0)
    .link(
      'http_readiness', () => probeScheme === 'HTTP',
      'connection_readiness', () => Boolean(probeTarget),
      'deny_readiness'
    )

    //
    // HTTP server for readiness probe
    //
    .pipeline('http_readiness')
    .demuxHTTP('message_readiness')

    //
    // rewrite request URL
    //
    .pipeline('message_readiness')
    .handleMessageStart(
      msg => (
        msg.head.path === '/osm-readiness-probe' && (msg.head.path = '/readiness'),
        probePath && (msg.head.path = probePath)
      )
    )
    .muxHTTP('connection_readiness', probeTarget)

    //
    // connect to the app port
    //
    .pipeline('connection_readiness')
    .connect(() => probeTarget)

    //
    // No target detected, access denied.
    //
    .pipeline('deny_readiness')
    .replaceStreamStart(
      new StreamEnd('ConnectionReset')
    )

    // startup probe
    .listen(probeScheme ? 15903 : 0)
    .link(
      'http_startup', () => probeScheme === 'HTTP',
      'connection_startup', () => Boolean(probeTarget),
      'deny_startup'
    )
    //
    // HTTP server for startup probe
    //
    .pipeline('http_startup')
    .demuxHTTP('message_startup')

    //
    // rewrite request URL
    //
    .pipeline('message_startup')
    .handleMessageStart(
      msg => (
        msg.head.path === '/osm-startup-probe' && (msg.head.path = '/startup'),
        probePath && (msg.head.path = probePath)
      )
    )
    .muxHTTP('connection_startup', probeTarget)

    //
    // connect to the app port
    //
    .pipeline('connection_startup')
    .connect(() => probeTarget)

    //
    // No target detected, access denied.
    //
    .pipeline('deny_startup')
    .replaceStreamStart(
      new StreamEnd('ConnectionReset')
    )

    // Prometheus collects metrics
    .listen(15010)
    .link('http_prometheus')

    //
    // HTTP server for Prometheus collection metrics
    //
    .pipeline('http_prometheus')
    .demuxHTTP('message_prometheus')

    //
    // Forward request to PIPY /metrics
    //
    .pipeline('message_prometheus')
    .handleMessageStart(
      msg => (
        (msg.head.path === '/stats/prometheus' && (msg.head.path = '/metrics')) || (msg.head.path = '/stats' + msg.head.path)
      )
    )
    .muxHTTP('connection_prometheus', () => prometheusTarget)

    //
    // PIPY admin: '127.0.0.1:6060'
    //
    .pipeline('connection_prometheus')
    .connect(() => prometheusTarget)

    .listen(15000)
    .serveHTTP(
      msg =>
      http.File.from('pipy.json').toMessage(msg.head.headers['accept-encoding'])
    )

  )
)())(JSON.decode(pipy.load('pipy.json')))
