((
  {
    namespace,
    kind,
    name,
    pod,
    upstreamResponseTotal,
    upstreamResponseCode,
  } = pipy.solve('modules/metrics.js'),

) => (

pipy()

.import({
  __clusterName: 'inbound-main',
})

.pipeline()
.chain()
.handleMessageStart(
  (msg) => (
    ((headers) => (
      (headers = msg?.head?.headers) && (() => (
        headers['osm-stats-namespace'] = namespace,
        headers['osm-stats-kind'] = kind,
        headers['osm-stats-name'] = name,
        headers['osm-stats-pod'] = pod,
        upstreamResponseTotal.withLabels(namespace, kind, name, pod, __clusterName).increase(),
        upstreamResponseCode.withLabels(msg?.head?.status?.toString().charAt(0), namespace, kind, name, pod, __clusterName).increase()
      ))()
    ))()
  )
)

))()