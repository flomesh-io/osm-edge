// version: '2022.07.07'
((
  {
    config,
    debugLogLevel,
    namespace,
    kind,
    name,
    pod,
    tlsCertChain,
    tlsPrivateKey,
    tlsIssuingCA,
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
    logZipkin,
    metrics
  } = pipy.solve('config.js')
) => (

  // Turn On Activity Metrics
  metrics.serverLiveGauge.increase(),

  metrics.tracingAddress &&
  (logZipkin = new logging.JSONLogger('zipkin').toHTTP('http://' + metrics.tracingAddress + metrics.tracingEndpoint, {
    batch: {
      prefix: '[',
      postfix: ']',
      separator: ','
    },
    headers: {
      'Host': metrics.tracingAddress,
      'Content-Type': 'application/json',
    }
  }).log),

  pipy({
    _inMatch: null,
    _inTarget: null,
    _inSessionControl: null,
    _ingressMode: null,
    _inZipkinStruct: {},
    _localClusterName: null,
    _outIP: null,
    _outPort: null,
    _outMatch: null,
    _outTarget: null,
    _outSessionControl: null,
    _egressMode: null,
    _outZipkinStruct: {},
    _upstreamClusterName: null,
    _outRequestTime: 0,
    _egressTargetMap: {}
  })

    //
    // inbound
    //
    .listen(config?.Inbound?.TrafficMatches ? 15003 : 0, {
      'transparent': true
    })
    .onStart(
      () => (
        (() => (
          // Find a match by destination port
          _inMatch = (
            allowedEndpoints?.[__inbound.remoteAddress || '127.0.0.1'] &&
            inTrafficMatches?.[__inbound.destinationPort || 0]
          ),

          // Check client address against the whitelist
          _inMatch?.AllowedEndpoints &&
          _inMatch.AllowedEndpoints[__inbound.remoteAddress] === undefined && (
            _inMatch = null
          ),

          // INGRESS mode
          _ingressMode = _inMatch?.SourceIPRanges?.find?.(e => e.contains(__inbound.remoteAddress)),

          // Layer 4 load balance
          _inTarget = (
            (
              // Allow?
              _inMatch &&
              _inMatch.Protocol !== 'http' && _inMatch.Protocol !== 'grpc'
            ) && (
              // Load balance
              inClustersConfigs?.[
                _localClusterName = _inMatch.TargetClusters?.next?.()?.id
              ]?.next?.()
            )
          ),

          // Session termination control
          _inSessionControl = {
            close: false
          },

          debugLogLevel && (
            console.log('inbound _inMatch: ', _inMatch) ||
            console.log('inbound _inTarget: ', _inTarget?.id) ||
            console.log('inbound protocol: ', _inMatch?.Protocol) ||
            console.log('inbound acceptTLS: ', Boolean(tlsCertChain))
          )
        ))(),
        !_inMatch || (_inTarget && _inMatch.Protocol !== 'http' && _inMatch.Protocol !== 'grpc') ? new Data : null
      )
    )
    .branch(
      () => Boolean(tlsCertChain) && Boolean(_inMatch) && !Boolean(_ingressMode), $ => $
        .acceptTLS({
          certificate: () => ({
            cert: new crypto.Certificate(tlsCertChain),
            key: new crypto.PrivateKey(tlsPrivateKey),
          }),
          trusted: (!tlsIssuingCA && []) || [
            new crypto.Certificate(tlsIssuingCA),
          ]
        }).to('recv-inbound-tcp'),
      'recv-inbound-tcp'
    )

    //
    // check inbound protocol
    //
    .pipeline('recv-inbound-tcp')
    .branch(
      () => _inMatch?.Protocol === 'http' || _inMatch?.Protocol === 'grpc', $ => $
        .demuxHTTP().to($ => $
          .replaceMessageStart(
            evt => _inSessionControl.close ? new StreamEnd : evt
          ).link('recv-inbound-http')
        ),
      () => Boolean(_inTarget), 'send-inbound-tcp',
      $ => $
        .replaceStreamStart(
          new StreamEnd('ConnectionReset')
        )
    )

    //
    // Analyze inbound HTTP request headers and match routes
    //
    .pipeline('recv-inbound-http')
    .handleMessageStart(
      (msg) => (
        ((service, match, headers) => (
          headers = msg.head.headers,

          // INGRESS mode
          // When found in SourceIPRanges, service is '*'
          _ingressMode && (service = '*'),

          // Find the service
          // When serviceidentity is present, service is headers.host
          !service && (service = (headers.serviceidentity && _inMatch?.HttpHostPort2Service?.[headers.host])),

          // Find a match by the service's route rules
          match = _inMatch.HttpServiceRouteRules?.[service]?.find?.(o => (
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
              _localClusterName = match?.TargetClusters?.next?.()?.id
            ]?.next?.()
          ),

          // Close sessions from any HTTP proxies
          !_inTarget && headers['x-forwarded-for'] && (
            _inSessionControl.close = true
          ),

          // Initialize ZipKin tracing data
          logZipkin && (() => (
            _inZipkinStruct.data = metrics.funcMakeZipKinData(name, msg, headers, _localClusterName, 'SERVER', true),
            _inZipkinStruct.requestSize = 0,
            _inZipkinStruct.responseSize = 0
          ))(),

          debugLogLevel && (
            console.log('inbound path: ', msg.head.path) ||
            console.log('inbound headers: ', msg.head.headers) ||
            console.log('inbound service: ', service) ||
            console.log('inbound match: ', match) ||
            console.log('inbound _inTarget: ', _inTarget?.id)
          )
        ))()
      )
    )
    .branch(
      () => Boolean(_inTarget) && _inMatch?.Protocol === 'grpc', $ => $
        .muxHTTP('send-inbound-tcp', () => _inTarget, {
          version: 2
        }),
      () => Boolean(_inTarget), $ => $
        .muxHTTP('send-inbound-tcp', () => _inTarget),
      $ => $
        .replaceMessage(
          new Message({
            status: 403
          }, 'Access denied')
        )
    )
    .handleMessageStart(
      (msg) => (
        ((headers) => (
          (headers = msg?.head?.headers) && (() => (
            headers['osm-stats-namespace'] = namespace,
            headers['osm-stats-kind'] = kind,
            headers['osm-stats-name'] = name,
            headers['osm-stats-pod'] = pod,
            metrics.upstreamResponseTotal.withLabels(namespace, kind, name, pod, _localClusterName).increase(),
            metrics.upstreamResponseCode.withLabels(msg?.head?.status?.toString().charAt(0), namespace, kind, name, pod, _localClusterName).increase(),

            _inZipkinStruct.data && (() => (
              _inZipkinStruct.data.tags['peer.address'] = _inTarget?.id,
              _inZipkinStruct.data.tags['http.status_code'] = msg?.head?.status?.toString?.(),
              _inZipkinStruct.data.tags['request_size'] = _inZipkinStruct.requestSize.toString(),
              _inZipkinStruct.data.tags['response_size'] = _inZipkinStruct.responseSize.toString(),
              _inZipkinStruct.data['duration'] = Date.now() * 1000 - _inZipkinStruct.data['timestamp'],
              logZipkin(_inZipkinStruct.data)
            ))(),

            debugLogLevel && console.log('_inZipkinStruct : ', _inZipkinStruct.data)
          ))()
        ))()
      )
    )

    //
    // Connect to local service
    //
    .pipeline('send-inbound-tcp')
    .onStart(
      () => (
        metrics.activeConnectionGauge.withLabels(_localClusterName).increase()
      )
    )
    .onEnd(
      () => (
        metrics.activeConnectionGauge.withLabels(_localClusterName).decrease()
      )
    )
    .handleData(
      (data) => (
        _inZipkinStruct.requestSize += data.size,
        metrics.sendBytesTotalCounter.withLabels(_localClusterName).increase(data.size)
      )
    )
    .connect(
      () => _inTarget?.id
    )
    .handleData(
      (data) => (
        _inZipkinStruct.responseSize += data.size,
        metrics.receiveBytesTotalCounter.withLabels(_localClusterName).increase(data.size)
      )
    )

    //
    // outbound
    //
    .listen(config?.Outbound || config?.Spec?.Traffic?.EnableEgress ? 15001 : 0, {
      'transparent': true
    })
    .onStart(
      () => (
        ((target) => (
          // Upstream service port
          _outPort = (__inbound.destinationPort || 0),

          // Upstream service IP
          _outIP = (__inbound.destinationAddress || '127.0.0.1'),

          _outMatch = (outTrafficMatches && outTrafficMatches[_outPort] && (
            // Strict matching Destination IP address
            outTrafficMatches[_outPort].find?.(o => (o.DestinationIPRanges && o.DestinationIPRanges.find(e => e.contains(_outIP)))) ||
            // EGRESS mode - does not check the IP
            (_egressMode = true) && outTrafficMatches[_outPort].find?.(o => (!Boolean(o.DestinationIPRanges) &&
              (o.Protocol == 'http' || o.Protocol == 'https' || (o.Protocol == 'tcp' && o.AllowedEgressTraffic))))
          )),

          // Layer 4 load balance
          _outTarget = (
            (
              // Allow?
              _outMatch &&
              _outMatch.Protocol !== 'http' && _outMatch.Protocol !== 'grpc'
            ) && (
              // Load balance
              outClustersConfigs?.[
                _upstreamClusterName = _outMatch.TargetClusters?.next?.()?.id
              ]?.next?.()
            )
          ),

          // EGRESS mode
          !Boolean(_outTarget) && (specEnableEgress || _outMatch?.AllowedEgressTraffic) && (_outMatch?.Protocol !== 'http') && (
            target = _outIP + ':' + _outPort,
            _upstreamClusterName = target,
            !_egressTargetMap[target] && (_egressTargetMap[target] = new algo.RoundRobinLoadBalancer({
              [target]: 100
            })),
            _outTarget = _egressTargetMap[target].next(),
            _egressMode = true
          ),

          _outSessionControl = {
            close: false
          },

          debugLogLevel && (
            console.log('outbound _outMatch: ', _outMatch) ||
            console.log('outbound _outTarget: ', _outTarget?.id) ||
            console.log('outbound protocol: ', _outMatch?.Protocol)
          )
        ))(),
        _outTarget && _outMatch?.Protocol !== 'http' && _outMatch?.Protocol !== 'grpc' ? new Data : null
      )
    )
    .branch(
      () => _outMatch?.Protocol === 'http' || _outMatch?.Protocol === 'grpc', $ => $
        .demuxHTTP().to($ => $
          .replaceMessageStart(
            evt => _outSessionControl.close ? new StreamEnd : evt
          ).link('recv-outbound-http')
        ),
      () => Boolean(_outTarget), 'send-outbound-tcp',
      $ => $
        .replaceStreamStart(
          new StreamEnd('ConnectionReset')
        )
    )

    //
    // Analyze outbound HTTP request headers and match routes
    //
    .pipeline('recv-outbound-http')
    .handleMessageStart(
      (msg) => (
        ((service, route, match, target, headers) => (
          headers = msg.head.headers,

          service = _outMatch.HttpHostPort2Service?.[headers.host],

          // Find route by HTTP host
          route = service && _outMatch.HttpServiceRouteRules?.[service],

          // Find a match by the service's route rules
          match = route?.find?.(o => (
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
              _upstreamClusterName = match?.TargetClusters?.next?.()?.id
            ]?.next?.()
          ),

          // Add serviceidentity for request authentication
          _outTarget && (headers['serviceidentity'] = _outMatch.ServiceIdentity),

          // Add x-b3 tracing Headers
          _outTarget && metrics.funcTracingHeaders(namespace, kind, name, pod, headers, _outMatch?.Protocol),

          // Initialize ZipKin tracing data
          logZipkin && (() => (
            _outZipkinStruct.data = metrics.funcMakeZipKinData(name, msg, headers, _upstreamClusterName, 'CLIENT', false),
            _outZipkinStruct.requestSize = 0,
            _outZipkinStruct.responseSize = 0
          ))(),

          // EGRESS mode
          !_outTarget && (specEnableEgress || _outMatch?.AllowedEgressTraffic) && (
            target = _outIP + ':' + _outPort,
            _upstreamClusterName = target,
            !_egressTargetMap[target] && (_egressTargetMap[target] = new algo.RoundRobinLoadBalancer({
              [target]: 100
            })),
            _outTarget = _egressTargetMap[target].next(),
            _egressMode = true
          ),

          _outRequestTime = Date.now(),

          debugLogLevel && (
            console.log('outbound path: ', msg.head.path) ||
            console.log('outbound headers: ', msg.head.headers) ||
            console.log('outbound service: ', service) ||
            console.log('outbound route: ', route) ||
            console.log('outbound match: ', match) ||
            console.log('outbound _outTarget: ', _outTarget?.id)
          )
        ))()
      )
    )
    .branch(
      () => Boolean(_outTarget) && _outMatch?.Protocol === 'grpc', $ => $
        .muxHTTP('send-outbound-tcp', () => _outTarget, {
          version: 2
        }),
      () => Boolean(_outTarget), $ => $
        .muxHTTP('send-outbound-tcp', () => _outTarget),
      $ => $
        .replaceMessage(
          new Message({
            status: 403
          }, 'Access denied')
        )
    )
    .handleMessageStart(
      (msg) => (
        ((headers, d_namespace, d_kind, d_name, d_pod) => (
          headers = msg?.head?.headers,
          (d_namespace = headers?.['osm-stats-namespace']) && (delete headers['osm-stats-namespace']),
          (d_kind = headers?.['osm-stats-kind']) && (delete headers['osm-stats-kind']),
          (d_name = headers?.['osm-stats-name']) && (delete headers['osm-stats-name']),
          (d_pod = headers?.['osm-stats-pod']) && (delete headers['osm-stats-pod']),
          d_namespace && metrics.osmRequestDurationHist.withLabels(namespace, kind, name, pod, d_namespace, d_kind, d_name, d_pod).observe(Date.now() - _outRequestTime),
          metrics.upstreamCompletedCount.withLabels(_upstreamClusterName).increase(),
          msg?.head?.status && metrics.upstreamCodeCount.withLabels(msg.head.status, _upstreamClusterName).increase(),
          msg?.head?.status && metrics.upstreamCodeXCount.withLabels(msg.head.status.toString().charAt(0), _upstreamClusterName).increase(),
          metrics.upstreamResponseTotal.withLabels(namespace, kind, name, pod, _upstreamClusterName).increase(),
          msg?.head?.status && metrics.upstreamResponseCode.withLabels(msg.head.status.toString().charAt(0), namespace, kind, name, pod, _upstreamClusterName).increase(),

          _outZipkinStruct.data && (() => (
            _outZipkinStruct.data.tags['peer.address'] = _outTarget?.id,
            _outZipkinStruct.data.tags['http.status_code'] = msg?.head?.status?.toString?.(),
            _outZipkinStruct.data.tags['request_size'] = _outZipkinStruct.requestSize.toString(),
            _outZipkinStruct.data.tags['response_size'] = _outZipkinStruct.responseSize.toString(),
            _outZipkinStruct.data['duration'] = Date.now() * 1000 - _outZipkinStruct.data['timestamp'],
            logZipkin(_outZipkinStruct.data)
          ))(),

          debugLogLevel && console.log('_outZipkinStruct : ', _outZipkinStruct.data)
        ))()
      )
    )

    //
    // Connect to upstream service
    //
    .pipeline('send-outbound-tcp')
    .onStart(
      () => (
        metrics.activeConnectionGauge.withLabels(_upstreamClusterName).increase()
      )
    )
    .onEnd(
      () => (
        metrics.activeConnectionGauge.withLabels(_upstreamClusterName).decrease()
      )
    )
    .handleData(
      (data) => (
        _outZipkinStruct.requestSize += data.size,
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
        _outZipkinStruct.responseSize += data.size,
        metrics.receiveBytesTotalCounter.withLabels(_upstreamClusterName).increase(data.size)
      )
    )

    //
    // liveness probe
    //
    .listen(probeScheme ? 15901 : 0)
    .branch(
      () => probeScheme === 'HTTP', $ => $
        .demuxHTTP().to($ => $
          .handleMessageStart(
            msg => (
              msg.head.path === '/osm-liveness-probe' && (msg.head.path = '/liveness'),
              probePath && (msg.head.path = probePath)
            )
          )
          .muxHTTP(() => probeTarget).to($ => $
            .connect(() => probeTarget)
          )
        ),
      () => Boolean(probeTarget), $ => $
        .connect(() => probeTarget),
      $ => $
        .replaceStreamStart(
          new StreamEnd('ConnectionReset')
        )
    )

    //
    // readiness probe
    //
    .listen(probeScheme ? 15902 : 0)
    .branch(
      () => probeScheme === 'HTTP', $ => $
        .demuxHTTP().to($ => $
          .handleMessageStart(
            msg => (
              msg.head.path === '/osm-readiness-probe' && (msg.head.path = '/readiness'),
              probePath && (msg.head.path = probePath)
            )
          )
          .muxHTTP(() => probeTarget).to($ => $
            .connect(() => probeTarget)
          )
        ),
      () => Boolean(probeTarget), $ => $
        .connect(() => probeTarget),
      $ => $
        .replaceStreamStart(
          new StreamEnd('ConnectionReset')
        )
    )

    //
    // startup probe
    //
    .listen(probeScheme ? 15903 : 0)
    .branch(
      () => probeScheme === 'HTTP', $ => $
        .demuxHTTP().to($ => $
          .handleMessageStart(
            msg => (
              msg.head.path === '/osm-startup-probe' && (msg.head.path = '/startup'),
              probePath && (msg.head.path = probePath)
            )
          )
          .muxHTTP(() => probeTarget).to($ => $
            .connect(() => probeTarget)
          )
        ),
      () => Boolean(probeTarget), $ => $
        .connect(() => probeTarget),
      $ => $
        .replaceStreamStart(
          new StreamEnd('ConnectionReset')
        )
    )

    //
    // Prometheus collects metrics
    //
    .listen(15010)
    .demuxHTTP()
    .to($ => $
      .handleMessageStart(
        msg => (
          (msg.head.path === '/stats/prometheus' && (msg.head.path = '/metrics')) || (msg.head.path = '/stats' + msg.head.path)
        )
      )
      .muxHTTP(() => prometheusTarget)
      .to($ => $
        .connect(() => prometheusTarget)
      )
    )

    //
    // PIPY configuration file
    //
    .listen(15000)
    .serveHTTP(
      msg => http.File.from('pipy.json').toMessage(msg.head.headers['accept-encoding'])
    )

))()

