((
  {
    namespace,
    kind,
    name,
    pod,
  } = pipy.solve('config.js'),
  metrics = pipy.solve('metrics-init.js')) => (

  pipy({
    requestTime: null
  })

    .import({
      _flow: 'main',
      _localClusterName: 'inbound-classifier',
      _upstreamClusterName: 'outbound-classifier'
    })

    .pipeline()

    .handleMessageStart(
      () => (
        (_flow == 'outbound') && (
          requestTime = Date.now()
        )
      )
    )

    .chain()

    .handleMessageStart(
      (msg) => (
        (_flow === 'inbound') && (
          ((headers) => (
            (headers = msg?.head?.headers) && (() => (
              headers['osm-stats-namespace'] = namespace,
              headers['osm-stats-kind'] = kind,
              headers['osm-stats-name'] = name,
              headers['osm-stats-pod'] = pod,
              metrics.upstreamResponseTotal.withLabels(namespace, kind, name, pod, _localClusterName).increase(),
              metrics.upstreamResponseCode.withLabels(msg?.head?.status?.toString().charAt(0), namespace, kind, name, pod, _localClusterName).increase()
            ))()
          ))()
        ),
        (_flow === 'outbound') && (
          ((headers, d_namespace, d_kind, d_name, d_pod) => (
            headers = msg?.head?.headers,
            (d_namespace = headers?.['osm-stats-namespace']) && (delete headers['osm-stats-namespace']),
            (d_kind = headers?.['osm-stats-kind']) && (delete headers['osm-stats-kind']),
            (d_name = headers?.['osm-stats-name']) && (delete headers['osm-stats-name']),
            (d_pod = headers?.['osm-stats-pod']) && (delete headers['osm-stats-pod']),
            d_namespace && metrics.osmRequestDurationHist.withLabels(namespace, kind, name, pod, d_namespace, d_kind, d_name, d_pod).observe(Date.now() - requestTime),
            metrics.upstreamCompletedCount.withLabels(_upstreamClusterName).increase(),
            msg?.head?.status && metrics.upstreamCodeCount.withLabels(msg.head.status, _upstreamClusterName).increase(),
            msg?.head?.status && metrics.upstreamCodeXCount.withLabels(msg.head.status.toString().charAt(0), _upstreamClusterName).increase(),
            metrics.upstreamResponseTotal.withLabels(namespace, kind, name, pod, _upstreamClusterName).increase(),
            msg?.head?.status && metrics.upstreamResponseCode.withLabels(msg.head.status.toString().charAt(0), namespace, kind, name, pod, _upstreamClusterName).increase()
          ))()
        )
      )
    )

))()