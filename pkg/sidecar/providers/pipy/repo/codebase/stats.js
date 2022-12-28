((
  config = pipy.solve('config.js'),
  prometheusTarget = '127.0.0.1:6060',
) => (

pipy({
  _statsPath: null,
})

.pipeline('prometheus')
.demuxHTTP().to($ => $
  .handleMessageStart(
    msg => (
      (msg.head.path === '/stats/prometheus' && (msg.head.path = '/metrics')) || (msg.head.path = '/stats' + msg.head.path)
    )
  )
  .muxHTTP(() => prometheusTarget).to($ => $
    .connect(() => prometheusTarget)
  )
)

//
// osm proxy get stats
//
.pipeline('osm-stats')
.demuxHTTP().to($ => $
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
      .muxHTTP(() => prometheusTarget).to($ => $
        .connect(() => prometheusTarget)
      )
      .replaceMessage(
        (msg, out) => (
          out = msg?.body?.toString()?.split?.('\n') || [],
          out = out.filter(line => line.indexOf('peer') > 0 || line.indexOf('_retry') > 0),
          (_statsPath === '/clusters') && (out = out.filter(line => line.indexOf('_bucket') < 0)),
          out = out.map(
            s => s.indexOf('rq_retry') > 0 ? (
              (
                items = s.replace('sidecar_cluster_', '').split('{').join(',').split('"').join(',').split(' ').join(',').split(','),
              ) => 'cluster.' + items[2] + '.' + items[0] + ': ' + items[4]
            )() : s
          ),
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
)

))()