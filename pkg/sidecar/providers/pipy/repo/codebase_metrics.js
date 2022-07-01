(
  (metrics) => (

    metrics = {

      logZipkin: null,
      tracingAddress: os.env.TRACING_ADDRESS,
      tracingEndpoint: (os.env.TRACING_ENDPOINT || '/api/v2/spans'),

      sendBytesTotalCounter: new stats.Counter('sidecar_cluster_upstream_cx_tx_bytes_total', ['sidecar_cluster_name']),
      receiveBytesTotalCounter: new stats.Counter('sidecar_cluster_upstream_cx_rx_bytes_total', ['sidecar_cluster_name']),
      activeConnectionGauge: new stats.Gauge('sidecar_cluster_upstream_cx_active', ['sidecar_cluster_name']),
      upstreamCodeCount: new stats.Counter('sidecar_cluster_external_upstream_rq', ['sidecar_response_code', 'sidecar_cluster_name']),
      upstreamCodeXCount: new stats.Counter('sidecar_cluster_external_upstream_rq_xx', ['sidecar_response_code_class', 'sidecar_cluster_name']),
      upstreamCompletedCount: new stats.Counter('sidecar_cluster_external_upstream_rq_completed', ['sidecar_cluster_name']),
      upstreamResponseTotal: new stats.Counter('sidecar_cluster_upstream_rq_total',
        ['source_namespace', 'source_workload_kind', 'source_workload_name', 'source_workload_pod', 'sidecar_cluster_name']),
      upstreamResponseCode: new stats.Counter('sidecar_cluster_upstream_rq_xx',
        ['sidecar_response_code_class', 'source_namespace', 'source_workload_kind', 'source_workload_name', 'source_workload_pod', 'sidecar_cluster_name']),
      osmRequestDurationHist: new stats.Histogram('osm_request_duration_ms',
        [5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000, 30000, 60000, 300000, 600000, 1800000, 3600000, Infinity],
        ['source_namespace', 'source_kind', 'source_name', 'source_pod', 'destination_namespace', 'destination_kind', 'destination_name', 'destination_pod']),
      serverLiveGauge: new stats.Gauge('sidecar_server_live'),
      //// serverLiveGauge.increase(),
      // {{{ TBD begin
      destroyRemoteActiveCounter: new stats.Counter('sidecar_cluster_upstream_cx_destroy_remote_with_active_rq', ['sidecar_cluster_name']),
      destroyLocalActiveCounter: new stats.Counter('sidecar_cluster_upstream_cx_destroy_local_with_active_rq', ['sidecar_cluster_name']),
      connectTimeoutCounter: new stats.Counter('sidecar_cluster_upstream_cx_connect_timeout', ['sidecar_cluster_name']),
      pendingFailureEjectCounter: new stats.Counter('sidecar_cluster_upstream_rq_pending_failure_eject', ['sidecar_cluster_name']),
      pendingOverflowCounter: new stats.Counter('sidecar_cluster_upstream_rq_pending_overflow', ['sidecar_cluster_name']),
      requestTimeoutCounter: new stats.Counter('sidecar_cluster_upstream_rq_timeout', ['sidecar_cluster_name']),
      requestReceiveResetCounter: new stats.Counter('sidecar_cluster_upstream_rq_rx_reset', ['sidecar_cluster_name']),
      requestSendResetCounter: new stats.Counter('sidecar_cluster_upstream_rq_tx_reset', ['sidecar_cluster_name']),
      // }}} TBD end

    },

    metrics.tracingAddress &&
    (metrics.logZipkin = new logging.JSONLogger('zipkin').toHTTP('http://' + metrics.tracingAddress + metrics.tracingEndpoint, {
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

    metrics.funcInitClusterNameMetrics = (namespace, kind, name, pod, clusterName) => (
      metrics.upstreamResponseTotal.withLabels(namespace, kind, name, pod, clusterName).zero(),
      metrics.upstreamResponseCode.withLabels('5', namespace, kind, name, pod, clusterName).zero(),
      metrics.activeConnectionGauge.withLabels(clusterName).zero(),
      metrics.receiveBytesTotalCounter.withLabels(clusterName).zero(),
      metrics.sendBytesTotalCounter.withLabels(clusterName).zero(),
      metrics.connectTimeoutCounter.withLabels(clusterName).zero(),
      metrics.destroyLocalActiveCounter.withLabels(clusterName).zero(),
      metrics.destroyRemoteActiveCounter.withLabels(clusterName).zero(),
      metrics.pendingFailureEjectCounter.withLabels(clusterName).zero(),
      metrics.pendingOverflowCounter.withLabels(clusterName).zero(),
      metrics.requestTimeoutCounter.withLabels(clusterName).zero(),
      metrics.requestReceiveResetCounter.withLabels(clusterName).zero(),
      metrics.requestSendResetCounter.withLabels(clusterName).zero()
    ),

    metrics.funcTracingHeaders = (namespace, kind, name, pod, headers, proto, uuid, id) => (
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

    metrics.funcMakeZipKinData = (name, msg, headers, clusterName, kind, shared, data) => (
      data = {
        'traceId': headers?.['x-b3-traceid'] && headers['x-b3-traceid'].toString(),
        'id': headers?.['x-b3-spanid'] && headers['x-b3-spanid'].toString(),
        'name': headers?.host,
        'timestamp': Date.now() * 1000,
        'localEndpoint': {
          'port': 0,
          'ipv4': os.env.POD_IP || '',
          'serviceName': name,
        },
        'tags': {
          'component': 'proxy',
          'http.url': headers?.['x-forwarded-proto'] + '://' + headers?.host + msg?.head?.path,
          'http.method': msg?.head?.method,
          'node_id': os.env.POD_UID || '',
          'http.protocol': msg?.head?.protocol,
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

    metrics
  )
)()
