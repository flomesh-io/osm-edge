((
  config = pipy.solve('config.js'),
  probeScheme = config?.Spec?.Probes?.LivenessProbes?.[0]?.httpGet?.scheme,
) => pipy()

.export('main', {
  __flow: '',
})

.branch(
  Boolean(config?.Inbound?.TrafficMatches), (
    $=>$
    .listen(15003, { transparent: true })
    .onStart(() => (__flow = 'inbound', new Data))
    .use('modules/inbound-main.js')
  )
)

.branch(
  Boolean(config?.Outbound || config?.Spec?.Traffic?.EnableEgress), (
    $=>$
    .listen(15001, { transparent: true })
    .onStart(() => (__flow = 'outbound', new Data))
    .use('modules/outbound-main.js')
  )
)

.listen(probeScheme ? 15901 : 0)
.use('probes.js', 'liveness')

.listen(probeScheme ? 15902 : 0)
.use('probes.js', 'readiness')

.listen(probeScheme ? 15903 : 0)
.use('probes.js', 'startup')

.listen(15010)
.use('stats.js', 'prometheus')

.listen(':::15000')
.use('stats.js', 'osm-stats')

)()
