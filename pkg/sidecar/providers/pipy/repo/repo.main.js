// version: '2022.06.27-rc3'
(config => (
  (
    debugLogLevel,
    namespace,
    kind,
    name,
    pod,

    // {{{ metrics begin
    serverLiveGauge,
    activeConnectionGauge,
    sendBytesTotalCounter,
    receiveBytesTotalCounter,
    upstreamResponseTotal,
    upstreamResponseCode,
    osmRequestDurationHist,
    upstreamCodeCount,
    upstreamCodeXCount,
    upstreamCompletedCount,
    funcInitClusterNameMetrics,
    destroyRemoteActiveCounter, // zero - To Be Determined
    destroyLocalActiveCounter, // zero - To Be Determined
    connectTimeoutCounter, // zero - To Be Determined
    pendingFailureEjectCounter, // zero - To Be Determined
    pendingOverflowCounter, // zero - To Be Determined
    requestTimeoutCounter, // zero - To Be Determined
    requestReceiveResetCounter, // zero - To Be Determined
    requestSendResetCounter, // zero - To Be Determined
    // }}} metrics end

    funcTracingHeaders,
    funcMakeZipKinData,
    tlsCertChain,
    tlsPrivateKey,
    tlsIssuingCA,

    dummyCertChain = "-----BEGIN CERTIFICATE-----\n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\nPCIiU5QthEGPPoerlJc/2P8F0VvBdnkwO8Wy/3A=\n-----END CERTIFICATE-----\n",
    dummyPrivateKey = "-----BEGIN PRIVATE KEY-----\n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\nE4QvevN7kg40ZiPcrYVGtYsFslf5Ayc=\n-----END PRIVATE KEY-----\n",

    tracingAddress,
    tracingEndpoint,
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
    debugLogLevel = (config?.Spec?.SidecarLogLevel === 'debug'),
    namespace = (os.env.POD_NAMESPACE || 'default'),
    kind = (os.env.POD_CONTROLLER_KIND || 'Deployment'),
    name = (os.env.SERVICE_ACCOUNT || ''),
    pod = (os.env.POD_NAME || ''),
    tracingAddress = (os.env.TRACING_ADDRESS || 'jaeger.osm-system.svc.cluster.local:9411'),
    tracingEndpoint = (os.env.TRACING_ENDPOINT || '/api/v2/spans'),

    tlsCertChain = config?.Certificate?.CertChain,
    tlsPrivateKey = config?.Certificate?.PrivateKey,
    tlsIssuingCA = config?.Certificate?.IssuingCA,

    sendBytesTotalCounter = new stats.Counter('envoy_cluster_upstream_cx_tx_bytes_total', ['envoy_cluster_name']),
    receiveBytesTotalCounter = new stats.Counter('envoy_cluster_upstream_cx_rx_bytes_total', ['envoy_cluster_name']),
    activeConnectionGauge = new stats.Gauge('envoy_cluster_upstream_cx_active', ['envoy_cluster_name']),
    upstreamCodeCount = new stats.Counter('envoy_cluster_external_upstream_rq', ['envoy_response_code', 'envoy_cluster_name']),
    upstreamCodeXCount = new stats.Counter('envoy_cluster_external_upstream_rq_xx', ['envoy_response_code_class', 'envoy_cluster_name']),
    upstreamCompletedCount = new stats.Counter('envoy_cluster_external_upstream_rq_completed', ['envoy_cluster_name']),
    upstreamResponseTotal = new stats.Counter('envoy_cluster_upstream_rq_total',
      ['source_namespace', 'source_workload_kind', 'source_workload_name', 'source_workload_pod', 'envoy_cluster_name']),
    upstreamResponseCode = new stats.Counter('envoy_cluster_upstream_rq_xx',
      ['envoy_response_code_class', 'source_namespace', 'source_workload_kind', 'source_workload_name', 'source_workload_pod', 'envoy_cluster_name']),
    osmRequestDurationHist = new stats.Histogram('osm_request_duration_ms',
      [5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000, 30000, 60000, 300000, 600000, 1800000, 3600000, Infinity],
      ['source_namespace', 'source_kind', 'source_name', 'source_pod', 'destination_namespace', 'destination_kind', 'destination_name', 'destination_pod']),
    serverLiveGauge = new stats.Gauge('envoy_server_live'),
    serverLiveGauge.increase(),
    // {{{ TBD begin
    destroyRemoteActiveCounter = new stats.Counter('envoy_cluster_upstream_cx_destroy_remote_with_active_rq', ['envoy_cluster_name']),
    destroyLocalActiveCounter = new stats.Counter('envoy_cluster_upstream_cx_destroy_local_with_active_rq', ['envoy_cluster_name']),
    connectTimeoutCounter = new stats.Counter('envoy_cluster_upstream_cx_connect_timeout', ['envoy_cluster_name']),
    pendingFailureEjectCounter = new stats.Counter('envoy_cluster_upstream_rq_pending_failure_eject', ['envoy_cluster_name']),
    pendingOverflowCounter = new stats.Counter('envoy_cluster_upstream_rq_pending_overflow', ['envoy_cluster_name']),
    requestTimeoutCounter = new stats.Counter('envoy_cluster_upstream_rq_timeout', ['envoy_cluster_name']),
    requestReceiveResetCounter = new stats.Counter('envoy_cluster_upstream_rq_rx_reset', ['envoy_cluster_name']),
    requestSendResetCounter = new stats.Counter('envoy_cluster_upstream_rq_tx_reset', ['envoy_cluster_name']),
    // }}} TBD end

    funcInitClusterNameMetrics = (clusterName) => (
      upstreamResponseTotal.withLabels(namespace, kind, name, pod, clusterName).zero(),
      upstreamResponseCode.withLabels('5', namespace, kind, name, pod, clusterName).zero(),
      activeConnectionGauge.withLabels(clusterName).zero(),
      receiveBytesTotalCounter.withLabels(clusterName).zero(),
      sendBytesTotalCounter.withLabels(clusterName).zero(),

      connectTimeoutCounter.withLabels(clusterName).zero(),
      destroyLocalActiveCounter.withLabels(clusterName).zero(),
      destroyRemoteActiveCounter.withLabels(clusterName).zero(),
      pendingFailureEjectCounter.withLabels(clusterName).zero(),
      pendingOverflowCounter.withLabels(clusterName).zero(),
      requestTimeoutCounter.withLabels(clusterName).zero(),
      requestReceiveResetCounter.withLabels(clusterName).zero(),
      requestSendResetCounter.withLabels(clusterName).zero()
    ),

    funcTracingHeaders = (headers, proto, uuid, id) => (
      uuid = algo.uuid(),
      id = algo.hash(uuid),
      proto && (headers['x-forwarded-proto'] = proto),
      headers['x-b3-spanid'] &&
      (headers['x-b3-parentspanid'] = headers['x-b3-spanid']) &&
      (headers['x-b3-spanid'] = id),
      !headers['x-b3-traceid'] &&
      (headers['x-b3-traceid'] = id) &&
      (headers['x-b3-spanid'] = id) &&
      (headers['x-b3-sampled'] = '1'),
      !headers['x-request-id'] && (headers['x-request-id'] = uuid),
      headers['osm-stats-namespace'] = namespace,
      headers['osm-stats-kind'] = kind,
      headers['osm-stats-name'] = name,
      headers['osm-stats-pod'] = pod
    ),

    funcMakeZipKinData = (msg, headers, clusterName, kind, shared, data) => (
      data = {
        'traceId': headers?.['x-b3-traceid'] && headers['x-b3-traceid'].toString(),
        'id': headers?.['x-b3-spanid'] && headers['x-b3-spanid'].toString(),
        'name': headers.host,
        'timestamp': Date.now() * 1000,
        'localEndpoint': {
          'port': 0,
          'ipv4': os.env.POD_IP || '',
          'serviceName': name,
        },
        'tags': {
          'component': 'proxy',
          'http.url': headers?.['x-forwarded-proto'] + '://' + headers.host + msg.head.path,
          'http.method': msg.head.method,
          'node_id': os.env.POD_UID || '',
          'http.protocol': msg.head.protocol,
          'guid:x-request-id': headers?.['x-request-id'],
          'user_agent': headers?.['user-agent'],
          'upstream_cluster': clusterName
        },
        'annotations': []
      },
      headers['x-b3-parentspanid'] && (data['parentId'] = headers['x-b3-parentspanid']),
      data['kind'] = kind,
      shared && (data['shared'] = shared),
      data.tags['request_size'] = '0',
      data.tags['response_size'] = '0',
      data.tags['http.status_code'] = '502',
      data.tags['peer.address'] = '',
      data['duration'] = 0,
      data
    ),

    funcHttpServiceRouteRules = json => (
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
            SourceIPRanges_: match?.SourceIPRanges, // for debugLogLevel
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
          k, (funcInitClusterNameMetrics(k), new algo.RoundRobinLoadBalancer(v))
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
          k, (funcInitClusterNameMetrics(k), new algo.RoundRobinLoadBalancer(v))
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
      _inMatch: undefined,
      _inTarget: undefined,
      _inSessionControl: null,
      _service: undefined,
      _outIP: undefined,
      _outPort: undefined,
      _outMatch: undefined,
      _outTarget: undefined,
      _outSessionControl: null,
      _egressMode: null,

      _outRequestTime: 0,
      _egressTargetMap: {},
      _localClusterName: undefined,
      _upstreamClusterName: undefined,
      _inZipkinData: null,
      _outZipkinData: null
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
        _service = _inMatch?.SourceIPRanges?.find?.(e => e.contains(__inbound.remoteAddress)),

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

    .link('inbound_tls', () => Boolean(tlsCertChain) && !Boolean(_service),
      'inbound_tls_offloaded')
    .pipeline('inbound_tls')
    .acceptTLS(
      'inbound_tls_offloaded', {
        certificate: {
          cert: new crypto.Certificate(tlsCertChain || dummyCertChain),
          key: new crypto.PrivateKey(tlsPrivateKey || dummyPrivateKey),
        },
        trusted: [
          new crypto.Certificate(tlsIssuingCA || dummyCertChain),
        ]
      }
    )
    .pipeline('inbound_tls_offloaded')

    .link(
      'http_in', () => _inMatch?.Protocol === 'http' || _inMatch?.Protocol === 'grpc',
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

          // INGRESS mode
          // When found in SourceIPRanges, service is '*'
          _service && (service = '*'),

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
          //// _inZipkinData = funcMakeZipKinData(msg, headers, _localClusterName, 'SERVER', true),

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
    .link(
      'request_in2', () => Boolean(_inTarget) && _inMatch?.Protocol === 'grpc',
      'request_in', () => Boolean(_inTarget),
      'deny_in_http'
    )

    //
    // Multiplexing access to HTTP/2 service
    //
    .pipeline('request_in2')
    .muxHTTP(
      'connection_in', () => _inTarget, {
        version: 2
      }
    )
    .link('local_response')

    //
    // Multiplexing access to HTTP service
    //
    .pipeline('request_in')
    .muxHTTP(
      'connection_in', () => _inTarget
    )
    .link('local_response')

    //
    // Connect to local service
    //
    .pipeline('connection_in')
    .handleData(
      (data) => (
        sendBytesTotalCounter.withLabels(_localClusterName).increase(data.size),
        _inZipkinData && (_inZipkinData.tags['request_size'] = data.size.toString())
      )
    )
    .handleStreamStart(
      () => (
        activeConnectionGauge.withLabels(_localClusterName).increase()
      )
    )
    .handleStreamEnd(
      () => (
        activeConnectionGauge.withLabels(_localClusterName).decrease()
      )
    )
    .connect(
      () => _inTarget?.id
    )
    .handleData(
      (data) => (
        receiveBytesTotalCounter.withLabels(_localClusterName).increase(data.size),
        _inZipkinData && (() => (
          _inZipkinData['duration'] = Date.now() * 1000 - _inZipkinData['timestamp'],
          _inZipkinData.tags['response_size'] = data.size.toString(),
          _inZipkinData.tags['peer.address'] = _inTarget.id
        ))()
      )
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

    //
    // local response
    //
    .pipeline('local_response')
    .handleMessageStart(
      (msg) => (
        ((headers) => (
          (headers = msg?.head?.headers) && (() => (
            headers['osm-stats-namespace'] = namespace,
            headers['osm-stats-kind'] = kind,
            headers['osm-stats-name'] = name,
            headers['osm-stats-pod'] = pod,

            upstreamResponseTotal.withLabels(namespace, kind, name, pod, _localClusterName).increase(),
            upstreamResponseCode.withLabels(msg?.head?.status?.toString().charAt(0), namespace, kind, name, pod, _localClusterName).increase(),

            _inZipkinData && (_inZipkinData.tags['http.status_code'] = msg?.head?.status?.toString()),
            debugLogLevel && console.log('_inZipkinData: ', _inZipkinData)
          ))()
        ))()
      )
    )
    .link('log_local_response', () => Boolean(_inZipkinData), '')

    //
    // sub-pipeline
    //
    .pipeline('log_local_response')
    .fork('fork_local_response')

    //
    // jaeger tracing for inbound
    //
    .pipeline('fork_local_response')
    .decompressHTTP()
    .replaceMessage(
      '4k',
      () => (
        new Message(
          JSON.encode([_inZipkinData]).push('\n')
        )
      )
    )
    .merge('send_tracing', () => '')

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
    .link(
      'http_out', () => _outMatch?.Protocol === 'http' || _outMatch?.Protocol === 'grpc',
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
          _outTarget && funcTracingHeaders(headers, _outMatch?.Protocol),

          // Initialize ZipKin tracing data
          //// _outZipkinData = funcMakeZipKinData(msg, headers, _upstreamClusterName, 'CLIENT', false),

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
    .link(
      'request_out2', () => Boolean(_outTarget) && _outMatch?.Protocol === 'grpc',
      'request_out', () => Boolean(_outTarget),
      'deny_out_http'
    )

    //
    // Multiplexing access to HTTP/2 service
    //
    .pipeline('request_out2')
    .muxHTTP(
      'connection_out', () => _outTarget, {
        version: 2
      }
    )
    .link('upstream_response')

    //
    // Multiplexing access to HTTP service
    //
    .pipeline('request_out')
    .muxHTTP(
      'connection_out', () => _outTarget
    )
    .link('upstream_response')

    //
    // Connect to upstream service
    //
    .pipeline('connection_out')
    .handleData(
      (data) => (
        sendBytesTotalCounter.withLabels(_upstreamClusterName).increase(data.size),
        _outZipkinData && (_outZipkinData.tags['request_size'] = data.size.toString())
      )
    )
    .handleStreamStart(
      () => (
        activeConnectionGauge.withLabels(_upstreamClusterName).increase()
      )
    )
    .handleStreamEnd(
      () => (
        activeConnectionGauge.withLabels(_upstreamClusterName).decrease()
      )
    )
    .link('tls_upstream_connect', () => (Boolean(tlsCertChain) && !Boolean(_egressMode)), 'upstream_connect')
    .pipeline('tls_upstream_connect')
    .connectTLS(
      'upstream_connect', {
        certificate: {
          cert: new crypto.Certificate(tlsCertChain || dummyCertChain),
          key: new crypto.PrivateKey(tlsPrivateKey || dummyPrivateKey),
        },
        trusted: [
          new crypto.Certificate(tlsIssuingCA || dummyCertChain),
        ]
      }
    )
    .pipeline('upstream_connect')
    .connect(
      () => _outTarget?.id
    )
    .handleData(
      (data) => (
        receiveBytesTotalCounter.withLabels(_upstreamClusterName).increase(data.size),
        _outZipkinData && (() => (
          _outZipkinData['duration'] = Date.now() * 1000 - _outZipkinData['timestamp'],
          _outZipkinData.tags['response_size'] = data.size.toString(),
          _outZipkinData.tags['peer.address'] = _outTarget.id
        ))()
      )
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

    //
    // upstram response
    //
    .pipeline('upstream_response')
    .handleMessageStart(
      (msg) => (
        ((headers, d_namespace, d_kind, d_name, d_pod) => (
          headers = msg?.head?.headers,
          (d_namespace = headers?.['osm-stats-namespace']) && (delete headers['osm-stats-namespace']),
          (d_kind = headers?.['osm-stats-kind']) && (delete headers['osm-stats-kind']),
          (d_name = headers?.['osm-stats-name']) && (delete headers['osm-stats-name']),
          (d_pod = headers?.['osm-stats-pod']) && (delete headers['osm-stats-pod']),

          d_namespace && osmRequestDurationHist.withLabels(namespace, kind, name, pod, d_namespace, d_kind, d_name, d_pod).observe(Date.now() - _outRequestTime),
          upstreamCompletedCount.withLabels(_upstreamClusterName).increase(),
          msg?.head?.status && upstreamCodeCount.withLabels(msg.head.status, _upstreamClusterName).increase(),
          msg?.head?.status && upstreamCodeXCount.withLabels(msg.head.status.toString().charAt(0), _upstreamClusterName).increase(),

          upstreamResponseTotal.withLabels(namespace, kind, name, pod, _upstreamClusterName).increase(),
          msg?.head?.status && upstreamResponseCode.withLabels(msg.head.status.toString().charAt(0), namespace, kind, name, pod, _upstreamClusterName).increase(),

          _outZipkinData && msg?.head?.status && (_outZipkinData.tags['http.status_code'] = msg.head.status.toString()),
          debugLogLevel && console.log('_outZipkinData: ', _outZipkinData)
        ))()
      )
    )
    .link('log_upstream_response', () => Boolean(_outZipkinData), '')

    //
    // sub-pipeline
    //
    .pipeline('log_upstream_response')
    .fork('fork_upstream_response')

    //
    // jaeger tracing for outbound
    //
    .pipeline('fork_upstream_response')
    .decompressHTTP()
    .replaceMessage(
      '4k',
      () => (
        new Message(
          JSON.encode([_outZipkinData]).push('\n')
        )
      )
    )
    .merge('send_tracing', () => '')

    //
    // send zipkin data to jaeger
    //
    .pipeline('send_tracing')
    .replaceMessageStart(
      () => new MessageStart({
        method: 'POST',
        path: tracingEndpoint,
        headers: {
          'Host': tracingAddress,
          'Content-Type': 'application/json',
        }
      })
    )
    .encodeHTTPRequest()
    .connect(
      () => tracingAddress, {
        bufferLimit: '8m',
      }
    )

    //
    // liveness probe
    //
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
    .muxHTTP('connection_liveness', () => probeTarget)

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

    //
    // readiness probe
    //
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
    .muxHTTP('connection_readiness', () => probeTarget)

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

    //
    // startup probe
    //
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
    .muxHTTP('connection_startup', () => probeTarget)

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

    //
    // PIPY configuration file
    //
    .listen(15000)
    .serveHTTP(
      msg =>
      http.File.from('pipy.json').toMessage(msg.head.headers['accept-encoding'])
    )

    .listen(14011)
    .serveHTTP(
      new Message('Hi, there!\n')
    )

  )
)())(JSON.decode(pipy.load('pipy.json')))
