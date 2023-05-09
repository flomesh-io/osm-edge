((
  config = pipy.solve('config.js'),
  probeScheme,
  probeTarget,
  probePath,
) => (
(
  // Initialize probeScheme, probeTarget, probePath
  config?.Inbound?.TrafficMatches && Object.entries(config.Inbound.TrafficMatches).map(
    ([port, match]) => (
      (match.Protocol === 'http') && (!probeTarget || !match.SourceIPRanges) && (probeTarget = '127.0.0.1:' + port)
    )
  ),
  config?.Spec?.Probes?.LivenessProbes && config.Spec.Probes.LivenessProbes[0]?.httpGet?.port == 15901 &&
  (probeScheme = config.Spec.Probes.LivenessProbes[0].httpGet.scheme) && !probeTarget &&
  ((probeScheme === 'HTTP' && (probeTarget = '127.0.0.1:80')) || (probeScheme === 'HTTPS' && (probeTarget = '127.0.0.1:443'))) &&
  (probePath = '/')
),

pipy()

.pipeline('liveness')
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

.pipeline('readiness')
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

.pipeline('startup')
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

))()