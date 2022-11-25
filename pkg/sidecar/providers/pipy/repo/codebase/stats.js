// version: '2022.09.19'
((
  {
    config,
    metrics,
    prometheusTarget
  } = pipy.solve('config.js')) => (

  pipy({
    _statsPath: null
  })

    //
    // osm proxy get stats
    //
    .pipeline()

    .handleMessageStart(
      msg => (
        _statsPath = msg.head.path,
        msg.head.path = '/metrics',
        delete msg.head.headers['accept-encoding'],
        (__inbound.remoteAddress === '127.0.0.1' && _statsPath === '/quitquitquit' && msg.head.method === 'POST') && (
          pipy.exit(),
          _statsPath = '/__exit'
        )
      )
    )
    .branch(
      () => (_statsPath === '/clusters' || _statsPath === '/stats'), $ => $
        .muxHTTP(() => prometheusTarget)
        .to($ => $
          .connect(() => prometheusTarget)
        )
        .replaceMessage(
          (msg, out) => (
            !(out = msg?.body?.toString()?.split?.('\n')) && (out = []),
            out = out.filter(line => line.indexOf('peer') > 0),
            (_statsPath === '/clusters') && (out = out.filter(line => line.indexOf('_bucket') < 0)),
            out = out.concat(Object.entries(metrics.sidecarInsideStats).map(
              ([k, v]) => (k + ': ' + v)
            )),
            new Message(out.join('\n'))
          )
        ),
      () => (_statsPath === '/listeners'), $ => $
        .replaceMessage(
          (msg) => (
            ((config?.Outbound || config?.Spec?.Traffic?.EnableEgress) && (msg = 'outbound-listener::0.0.0.0:15001\n')) || (msg = ''),
            (config?.Inbound?.TrafficMatches) && (msg += 'inbound-listener::0.0.0.0:15003\n'),
            msg += 'inbound-prometheus-listener::0.0.0.0:15010\n',
            new Message(msg)
          )
        ),
      () => (_statsPath === '/__exit'), $ => $
        .replaceMessage(
          new Message('DONE\n')
        ),
      () => (_statsPath === '/ready'), $ => $
        .replaceMessage(
          new Message('LIVE\n')
        ),
      () => (_statsPath === '/certs'), $ => $
        .replaceMessage(
          () => new Message(JSON.stringify(config.Certificate, null, 2))
        ),
      () => (_statsPath === '/config_dump'), $ => $
        .replaceMessage(
          msg => http.File.from('config.json').toMessage(msg.head.headers['accept-encoding'])
        ),
      $ => $
        .replaceMessage(
          new Message({
            status: 404
          }, 'Not Found\n')
        )
    )

))()