// version: '2022.06.30-rc3'
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
    probePath
  } = pipy.solve('config.js'),
  metrics = pipy.solve('metrics.js')
) => (

  // Turn On Activity Metrics
  metrics.serverLiveGauge.increase(),

  pipy({
    _inMatch: null,
    _inTarget: null,
    _inSessionControl: null,
    _ingressMode: null,
    _inZipkinData: null,
    _localClusterName: null,
    _outIP: null,
    _outPort: null,
    _outMatch: null,
    _outTarget: null,
    _outSessionControl: null,
    _egressMode: null,
    _outZipkinData: null,
    _upstreamClusterName: null,
    _outRequestTime: 0,
    _egressTargetMap: {}
  })

    //
    // inbound
    //
    .listen(config?.Inbound?.TrafficMatches ? 15003 : 0, {
      'transparent': true,
      'closeEOF': false
      // 'readTimeout': '5s'
    })
    .handleStreamStart(
      () => (
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
      )
    )
    .branch(
      () => Boolean(tlsCertChain) && !Boolean(_ingressMode), $ => $
        .acceptTLS({
          certificate: () => ({
            cert: new crypto.Certificate(tlsCertChain),
            key: new crypto.PrivateKey(tlsPrivateKey),
          }),
          trusted: (!tlsIssuingCA && []) || [
            new crypto.Certificate(tlsIssuingCA),
          ]
        }).to('recv-inbound-tcp'),
      () => true, 'recv-inbound-tcp'
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
      () => true, $ => $
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
          metrics.tracingAddress &&
          (_inZipkinData = metrics.funcMakeZipKinData(name, msg, headers, _localClusterName, 'SERVER', true)),

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
      () => true, $ => $
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
            _inZipkinData && (_inZipkinData.tags['http.status_code'] = msg?.head?.status?.toString()),
            debugLogLevel && console.log('_inZipkinData: ', _inZipkinData)
          ))()
        ))()
      )
    )
    .branch(
      () => Boolean(_inZipkinData), $ => $
        .fork().to($ => $
          .replaceMessage(
            '4k',
            () => (
              new Message(
                JSON.encode([_inZipkinData]).push('\n')
              )
            )
          )
          .merge('send-tracing', () => '')
        ),
      () => true, $ => $
    )

    //
    // Connect to local service
    //
    .pipeline('send-inbound-tcp')
    .handleData(
      (data) => (
        metrics.sendBytesTotalCounter.withLabels(_localClusterName).increase(data.size),
        _inZipkinData && (_inZipkinData.tags['request_size'] = data.size.toString())
      )
    )
    .handleStreamStart(
      () => (
        metrics.activeConnectionGauge.withLabels(_localClusterName).increase()
      )
    )
    .handleStreamEnd(
      () => (
        metrics.activeConnectionGauge.withLabels(_localClusterName).decrease()
      )
    )
    .connect(
      () => _inTarget?.id
    )
    .handleData(
      (data) => (
        metrics.receiveBytesTotalCounter.withLabels(_localClusterName).increase(data.size),
        _inZipkinData && (() => (
          _inZipkinData['duration'] = Date.now() * 1000 - _inZipkinData['timestamp'],
          _inZipkinData.tags['response_size'] = data.size.toString(),
          _inZipkinData.tags['peer.address'] = _inTarget.id
        ))()
      )
    )

    //
    // outbound
    //
    .listen(config?.Outbound || config?.Spec?.Traffic?.EnableEgress ? 15001 : 0, {
      'transparent': true,
      'closeEOF': false
      // 'readTimeout': '5s'
    })
    .handleStreamStart(
      (() => (
        (target) => (
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
        )
      ))()
    )
    .branch(
      () => _outMatch?.Protocol === 'http' || _outMatch?.Protocol === 'grpc', $ => $
        .demuxHTTP().to($ => $
          .replaceMessageStart(
            evt => _outSessionControl.close ? new StreamEnd : evt
          ).link('recv-outbound-http')
        ),
      () => Boolean(_outTarget), 'send-outbound-tcp',
      () => true, $ => $
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
          metrics.tracingAddress &&
          (_outZipkinData = metrics.funcMakeZipKinData(name, msg, headers, _upstreamClusterName, 'CLIENT', false)),

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
      () => true, $ => $
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
          _outZipkinData && msg?.head?.status && (_outZipkinData.tags['http.status_code'] = msg.head.status.toString()),
          debugLogLevel && console.log('_outZipkinData: ', _outZipkinData)
        ))()
      )
    )
    .branch(
      () => Boolean(_outZipkinData), $ => $
        .fork().to($ => $
          .replaceMessage(
            '4k',
            () => (
              new Message(
                JSON.encode([_outZipkinData]).push('\n')
              )
            )
          )
          .merge('send-tracing', () => '')
        ),
      () => true, $ => $
    )

    //
    // Connect to upstream service
    //
    .pipeline('send-outbound-tcp')
    .handleData(
      (data) => (
        metrics.sendBytesTotalCounter.withLabels(_upstreamClusterName).increase(data.size),
        _outZipkinData && (_outZipkinData.tags['request_size'] = data.size.toString())
      )
    )
    .handleStreamStart(
      () => (
        metrics.activeConnectionGauge.withLabels(_upstreamClusterName).increase()
      )
    )
    .handleStreamEnd(
      () => (
        metrics.activeConnectionGauge.withLabels(_upstreamClusterName).decrease()
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
      () => true, ($ => $
        .connect(() => _outTarget?.id)
    )
    )
    .handleData(
      (data) => (
        metrics.receiveBytesTotalCounter.withLabels(_upstreamClusterName).increase(data.size),
        _outZipkinData && (() => (
          _outZipkinData['duration'] = Date.now() * 1000 - _outZipkinData['timestamp'],
          _outZipkinData.tags['response_size'] = data.size.toString(),
          _outZipkinData.tags['peer.address'] = _outTarget.id
        ))()
      )
    )

    //
    // send zipkin data to jaeger
    //
    .pipeline('send-tracing')
    .replaceMessageStart(
      () => new MessageStart({
        method: 'POST',
        path: metrics.tracingEndpoint,
        headers: {
          'Host': metrics.tracingAddress,
          'Content-Type': 'application/json',
        }
      })
    )
    .encodeHTTPRequest()
    .connect(
      () => metrics.tracingAddress, {
      bufferLimit: '8m',
    }
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
      () => true, $ => $
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
      () => true, $ => $
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
      () => true, $ => $
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
