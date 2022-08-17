// version: '2022.07.28a'
pipy({
  _statsPath: null
})

  .import({
    config: 'main',
    prometheusTarget: 'main'
  })

  //
  // osm proxy get stats
  //
  .pipeline()

  .handleMessageStart(
    msg => (
      _statsPath = msg.head.path,
      msg.head.path = '/metrics',
      delete msg.head.headers['accept-encoding']
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
        msg => http.File.from('pipy.json').toMessage(msg.head.headers['accept-encoding'])
      ),
    $ => $
      .replaceMessage(
        new Message({
          status: 404
        }, 'Not Found\n')
      )
  )