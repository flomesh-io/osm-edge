(
  (
    {
      config,
      namespace,
      kind,
      name,
      pod,
    } = pipy.solve('config.js'),
    initClusterNameMetrics,
    metrics
  ) => (

    metrics = {

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

    initClusterNameMetrics = (namespace, kind, name, pod, clusterName) => (
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

    // pipy inside stats
    metrics.sidecarInsideStats = {},

    metrics.sidecarInsideStats['http_local_rate_limiter.http_local_rate_limit.rate_limited'] = 0,

    config?.Inbound?.TrafficMatches && Object.entries(config.Inbound.TrafficMatches).map(
      ([port, match]) => (
        metrics.sidecarInsideStats['local_rate_limit.inbound_' + namespace + '/' + pod.split('-')[0] + '_' + port + '_' + match?.Protocol + '.rate_limited'] = 0
      )
    ),

    config?.Inbound?.ClustersConfigs && Object.keys(config.Inbound.ClustersConfigs).map(
      key => (
        initClusterNameMetrics(namespace, kind, name, pod, key)
      )
    ),

    config?.Outbound?.ClustersConfigs && Object.keys(config.Outbound.ClustersConfigs).map(
      key => (
        initClusterNameMetrics(namespace, kind, name, pod, key),
        metrics.sidecarInsideStats['cluster.' + key + '.upstream_rq_retry'] = 0,
        metrics.sidecarInsideStats['cluster.' + key + '.upstream_rq_retry_backoff_exponential'] = 0,
        metrics.sidecarInsideStats['cluster.' + key + '.upstream_rq_retry_backoff_ratelimited'] = 0,
        metrics.sidecarInsideStats['cluster.' + key + '.upstream_rq_retry_limit_exceeded'] = 0,
        metrics.sidecarInsideStats['cluster.' + key + '.upstream_rq_retry_overflow'] = 0,
        metrics.sidecarInsideStats['cluster.' + key + '.upstream_rq_retry_success'] = 0,
        metrics.sidecarInsideStats['cluster.' + key + '.upstream_rq_pending_overflow'] = 0
      )
    ),

    // Turn On Activity Metrics
    metrics.serverLiveGauge.increase(),

    metrics
  )

)()