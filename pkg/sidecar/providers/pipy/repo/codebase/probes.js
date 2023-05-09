((
  config = pipy.solve('config.js'),
  inboundTarget,
  parseProbe,
  livenessProbe,
  readinessProbe,
  startupProbe,
) => (
  (
    config?.Inbound?.TrafficMatches && Object.entries(config.Inbound.TrafficMatches).map(
      ([port, match]) => (
        (match.Protocol === 'http') && (!inboundTarget || !match.SourceIPRanges) && (inboundTarget = '127.0.0.1:' + port)
      )
    ),

    parseProbe = (config, port, prefix) => (
      (path, target = inboundTarget, scheme) => (
        (config?.httpGet?.port === port) && (
          scheme = config?.httpGet?.scheme,
          config?.httpGet?.path?.startsWith?.(prefix) && (
            path = config.httpGet.path.substring(prefix.length),
            path?.split?.('/')?.[0] && (target = '127.0.0.1:' + path.split('/')[0]),
            !target && (
              (scheme === 'HTTP') && (target = '127.0.0.1:80'),
              (scheme === 'HTTPS') && (target = '127.0.0.1:443')
            ),
            path?.split?.('/')?.[1] ? (
              path = path.substring(path?.split?.('/')?.[0].length)
            ) : (path = '/')
          )
        ),
        { path, target, scheme }
      )
    )(),

    livenessProbe = parseProbe(config?.Spec?.Probes?.LivenessProbes?.[0], 15901, '/osm-liveness-probe/'),
    readinessProbe = parseProbe(config?.Spec?.Probes?.ReadinessProbes?.[0], 15902, '/osm-readiness-probe/'),
    startupProbe = parseProbe(config?.Spec?.Probes?.StartupProbes?.[0], 15903, '/osm-startup-probe/')

  ),

  pipy()

    .pipeline('liveness')
    .branch(
      () => livenessProbe?.scheme === 'HTTP', $=>$
        .demuxHTTP().to($=>$
          .handleMessageStart(
            msg => (
              msg.head.path = livenessProbe?.path
            )
          )
          .muxHTTP(() => livenessProbe?.target).to($=>$
            .connect(() => livenessProbe?.target)
          )
        ),
      () => Boolean(livenessProbe?.target), $=>$
        .connect(() => livenessProbe?.target),
      $=>$
        .replaceStreamStart(
          new StreamEnd('ConnectionReset')
        )
    )

    .pipeline('readiness')
    .branch(
      () => readinessProbe?.scheme === 'HTTP', $=>$
        .demuxHTTP().to($=>$
          .handleMessageStart(
            msg => (
              msg.head.path = readinessProbe?.path
            )
          )
          .muxHTTP(() => readinessProbe?.target).to($=>$
            .connect(() => readinessProbe?.target)
          )
        ),
      () => Boolean(readinessProbe?.target), $=>$
        .connect(() => readinessProbe?.target),
      $=>$
        .replaceStreamStart(
          new StreamEnd('ConnectionReset')
        )
    )

    .pipeline('startup')
    .branch(
      () => startupProbe?.scheme === 'HTTP', $=>$
        .demuxHTTP().to($=>$
          .handleMessageStart(
            msg => (
              msg.head.path = startupProbe?.path
            )
          )
          .muxHTTP(() => startupProbe?.target).to($=>$
            .connect(() => startupProbe?.target)
          )
        ),
      () => Boolean(startupProbe?.target), $=>$
        .connect(() => startupProbe?.target),
      $=>$
        .replaceStreamStart(
          new StreamEnd('ConnectionReset')
        )
    )

))()
