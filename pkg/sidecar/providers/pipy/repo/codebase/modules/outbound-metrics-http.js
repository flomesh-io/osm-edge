((
  {
    namespace,
    kind,
    name,
    pod,
    upstreamResponseTotal,
    upstreamResponseCode,
    osmRequestDurationHist,
    upstreamCompletedCount,
    upstreamCodeCount,
    upstreamCodeXCount,
  } = pipy.solve('modules/metrics.js'),
) => (

pipy({
  requestTime: null
})

.import({
  __clusterName: 'outbound-main'
})

.pipeline()
.handleMessageStart(
  () => (
    requestTime = Date.now()
  )
)
.chain()
.handleMessageStart(
  (msg) => (
    ((headers, d_namespace, d_kind, d_name, d_pod) => (
      headers = msg?.head?.headers,
      (d_namespace = headers?.['osm-stats-namespace']) && (delete headers['osm-stats-namespace']),
      (d_kind = headers?.['osm-stats-kind']) && (delete headers['osm-stats-kind']),
      (d_name = headers?.['osm-stats-name']) && (delete headers['osm-stats-name']),
      (d_pod = headers?.['osm-stats-pod']) && (delete headers['osm-stats-pod']),
      d_namespace && osmRequestDurationHist.withLabels(namespace, kind, name, pod, d_namespace, d_kind, d_name, d_pod).observe(Date.now() - requestTime),
      upstreamCompletedCount.withLabels(__clusterName).increase(),
      msg?.head?.status && upstreamCodeCount.withLabels(msg.head.status, __clusterName).increase(),
      msg?.head?.status && upstreamCodeXCount.withLabels(msg.head.status.toString().charAt(0), __clusterName).increase(),
      upstreamResponseTotal.withLabels(namespace, kind, name, pod, __clusterName).increase(),
      msg?.head?.status && upstreamResponseCode.withLabels(msg.head.status.toString().charAt(0), namespace, kind, name, pod, __clusterName).increase()
    ))()
  )
)

))()