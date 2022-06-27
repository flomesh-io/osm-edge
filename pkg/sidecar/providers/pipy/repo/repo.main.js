// version: '2022.06.26-dev'
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
      data.tags['http.status_code'] = '000',
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
      _outIP: undefined,
      _outPort: undefined,
      _outMatch: undefined,
      _outTarget: undefined,
      _outSessionControl: null,

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
    .acceptTLS(
      'inbound_tls_offloaded', {
        certificate: {
          cert: new crypto.Certificate(tlsCertChain),
          key: new crypto.PrivateKey(tlsPrivateKey),
        },
        trusted: [
          new crypto.Certificate(tlsIssuingCA),
        ]
      }
    )
    .pipeline('inbound_tls_offloaded')
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
          console.log('inbound protocol: ', _inMatch?.Protocol)
        )
      )
    )
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

          // Find the service
          service = (
            // When found in SourceIPRanges, service is '*'
            (_inMatch?.SourceIPRanges?.find?.(e => e.contains(__inbound.remoteAddress)) && '*') ||
            // When serviceidentity is present, service is headers.host
            (headers.serviceidentity && _inMatch?.HttpHostPort2Service?.[headers.host])
          ),

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
            outTrafficMatches[_outPort].find?.(o => (!Boolean(o.DestinationIPRanges) &&
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
            _outTarget = _egressTargetMap[target].next()
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
            _outTarget = _egressTargetMap[target].next()
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
    .connectTLS(
      'upstream_connect', {
        certificate: {
          cert: new crypto.Certificate(tlsCertChain),
          key: new crypto.PrivateKey(tlsPrivateKey),
        },
        trusted: [
          new crypto.Certificate(tlsIssuingCA),
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

          _outZipkinData && (_outZipkinData.tags['http.status_code'] = msg?.head?.status?.toString()),
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

  )
)())(JSON.decode(pipy.load('pipy.json')))
