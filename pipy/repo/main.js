(config =>

  pipy({
    _version: '2022.05.01',
    _targetCount: new stats.Counter('lb_target_cnt', ['target']),

    _specEnableEgress: config?.Spec?.Traffic?.EnableEgress,

    _inTrafficMatches: config?.Inbound?.TrafficMatches && Object.fromEntries(
      Object.entries(config.Inbound.TrafficMatches).map(
        ([k1, v1]) => [
          k1,
          Object.keys(v1).forEach(
            k2 => (
              (k2 == 'Protocol' && v1[k2] == 'http' && (!config._probeTarget || !v1.SourceIPRanges) && (config._probeTarget = '127.0.0.1:' + k1)),
              (k2 == 'TargetClusters') && v1[k2] && (v1['TargetClusters_'] = new algo.RoundRobinLoadBalancer(v1[k2])),
              (k2 == 'SourceIPRanges') && v1[k2] && (v1['SourceIPRanges_'] = []) && v1[k2].map(e => (v1['SourceIPRanges_'].push(new Netmask(e)))),
              (k2 == 'HttpServiceRouteRules') && v1[k2] && (
                Object.entries(v1[k2]).map(
                  ([k3, v3]) => [
                    k3,
                    Object.keys(v3).forEach(
                      k4 => (
                        Object.keys(v3[k4]).forEach(
                          k5 => (
                            (k5 == 'Methods') && (v3[k4]['Methods_'] = {}) && v3[k4][k5].map(
                              v6 => v3[k4]['Methods_'][v6] = true),
                            (k5 == 'Headers') && (v3[k4]['Headers_'] = {}) && Object.entries(v3[k4][k5]).map(
                              ([k6, v6]) => (v3[k4]['Headers_'][k6] = new RegExp(v6))),
                            (k5 == 'AllowedServices') && (v3[k4]['AllowedServices_'] = {}) && v3[k4][k5].map(
                              v6 => v3[k4]['AllowedServices_'][v6] = true),
                            (k5 == 'TargetClusters') && (v3[k4]['TargetClusters_'] = new algo.RoundRobinLoadBalancer(v3[k4][k5]))
                          )
                        ),
                        (v3[k4]['path'] == undefined) && (v3[k4]['path'] = new RegExp(k4))
                      )
                    ) || v3
                  ]
                )
              )
            )
          ) || v1
        ]
      )
    ),

    _inClustersConfigs: config?.Inbound?.ClustersConfigs && Object.fromEntries(
      Object.entries(
        config.Inbound.ClustersConfigs).map(
        ([k, v]) => [
          k, (new algo.RoundRobinLoadBalancer(v))
        ]
      )
    ),

    _outTrafficMatches: config?.Outbound?.TrafficMatches && config.Outbound.TrafficMatches.map(
      (o => (
        o.TargetClusters && (o['TargetClusters_'] = new algo.RoundRobinLoadBalancer(o.TargetClusters)),
        o.DestinationIPRanges && (o['DestinationIPRanges_'] = []) &&
        o.DestinationIPRanges.map(e => (o['DestinationIPRanges_'].push(new Netmask(e)))),
        o.HttpServiceRouteRules &&
        (Object.entries(o.HttpServiceRouteRules).map(
          ([k3, v3]) => [
            k3,
            Object.keys(v3).forEach(
              k4 => (
                Object.keys(v3[k4]).forEach(
                  k5 => (
                    (k5 == 'Methods') && (v3[k4]['Methods_'] = {}) && v3[k4][k5].map(
                      v6 => v3[k4]['Methods_'][v6] = true),
                    (k5 == 'Headers') && (v3[k4]['Headers_'] = {}) && Object.entries(v3[k4][k5]).map(
                      ([k6, v6]) => (v3[k4]['Headers_'][k6] = new RegExp(v6))),
                    (k5 == 'AllowedServices') && (v3[k4]['AllowedServices_'] = {}) && v3[k4][k5].map(
                      v6 => v3[k4]['AllowedServices_'][v6] = true),
                    (k5 == 'TargetClusters') && (v3[k4]['TargetClusters_'] = new algo.RoundRobinLoadBalancer(v3[k4][k5]))
                  )
                ),
                (v3[k4]['path'] == undefined) && (v3[k4]['path'] = new RegExp(k4))
              )
            ) || v3
          ]
        )),
        o
      ))
    ),

    _outClustersConfigs: config?.Outbound?.ClustersConfigs && Object.fromEntries(
      Object.entries(
        config.Outbound.ClustersConfigs).map(
        ([k, v]) => [
          k, (new algo.RoundRobinLoadBalancer(v))
        ]
      )
    ),

    _SpecProbes: config?.Spec?.Probes?.LivenessProbes && config.Spec.Probes.LivenessProbes[0]?.httpGet?.port == 15901 &&
      (config._probeScheme = config.Spec.Probes.LivenessProbes[0].httpGet.scheme) && !Boolean(config._probeTarget) &&
      ((config._probeScheme == 'HTTP' && (config._probeTarget = '127.0.0.1:80')) ||
        (config._probeScheme == 'HTTPS' && (config._probeTarget = '127.0.0.1:443'))) && (config._probePath = '/'),

    _AllowedEndpoints: config?.AllowedEndpoints,
    _prometheusTarget: '127.0.0.1:6060',

    _inPort: undefined,
    _inMatch: undefined,
    _inTarget: undefined,
    _inProtocol: undefined,
    _inSessionControl: null,
    _inClientIP: undefined,
    _inXForwardedFor: undefined,

    _outIP: undefined,
    _outPort: undefined,
    _outMatch: undefined,
    _outTarget: undefined,
    _outProtocol: undefined,
    _outSessionControl: null

  })

  // inbound
  .listen(config?.Inbound?.TrafficMatches ? 15003 : 0, {
    'transparent': true,
    'closeEOF': false
    // 'readTimeout': '5s'
  })
  .handleStreamStart(
    (_target) => (
      _inClientIP = __inbound.remoteAddress,
      _inPort = (__inbound?.destinationPort ? __inbound.destinationPort : '0'),
      _AllowedEndpoints && _AllowedEndpoints[_inClientIP] &&
      _inTrafficMatches && (_inMatch = _inTrafficMatches[_inPort]) &&
      (!Boolean(_inMatch['AllowedEndpoints']) || _inMatch['AllowedEndpoints'][_inClientIP]) &&
      (
        ((_inMatch['Protocol'] == 'http') && (_inProtocol = 'http')) ||
        ((_target = _inMatch['TargetClusters_'].select()) &&
          _inClustersConfigs[_target] && (_inTarget = _inClustersConfigs[_target].select()))
      ),
      _inSessionControl = {
        close: false
      },

      console.log('_inClientIP:' + _inClientIP),
      console.log('_inPort: ' + _inPort),
      console.log('_inProtocol: ' + _inProtocol),
      console.log('i_target: ' + _target),
      console.log('_inTarget: ' + _inTarget)

    )
  )
  .link(
    'http_in', () => _inProtocol == 'http',
    'connection_in', () => Boolean(_inTarget),
    'deny_in'
  )
  .pipeline('http_in')
  .demuxHTTP('inbound')
  .replaceMessageStart(
    evt => _inSessionControl.close ? new StreamEnd : evt
  )
  .pipeline('inbound')
  .handleMessageStart(
    (msg, _service, _route, _match, _target) => (
      msg.head?.headers['x-forwarded-for'] && (_inXForwardedFor = true),
      ((_inMatch.SourceIPRanges_ && _inMatch.SourceIPRanges_.find(e => e.contains(_inClientIP)) && (_service = "*")) ||
        (msg.head.headers?.host && (_service = _inMatch.HttpHostPort2Service[msg.head.headers.host]))) &&
      (_route = _inMatch.HttpServiceRouteRules[_service]) && (_service == "*" || Boolean(msg.head.headers['serviceidentity'])) &&
      (_match = Object.values(_route).find(o => (
        o.path.exec(msg.head.path) &&
        (!o.Methods_ || o.Methods_[msg.head.method] || o.Methods_['*']) &&
        (!o.AllowedServices_ || o.AllowedServices_[msg.head.headers['serviceidentity']] || o.AllowedServices_['*']) &&
        (!o.Headers_ || Object.entries(o.Headers_).every(
          h => msg.head.headers[h[0]] && h[1].exec(msg.head.headers[h[0]])
        ))))) &&
      (_target = _match.TargetClusters_.select()) && _inClustersConfigs[_target] && (_inTarget = _inClustersConfigs[_target].select()),

      console.log(msg.head),
      console.log('i_service: ' + _service),
      console.log('i_route: ' + _route),
      console.log('i_match: ' + _match),
      console.log('i_target: ' + _target),
      console.log('_inTarget: ' + _inTarget)

    )
  )
  .link(
    'request_in', () => Boolean(_inTarget),
    'deny_in_http'
  )
  .pipeline('request_in')
  .muxHTTP(
    'connection_in',
    () => _inTarget
  )
  .pipeline('connection_in')
  .connect(
    () => _inTarget
  )
  .pipeline('deny_in_http')
  .replaceMessage(
    msg => (
      _inSessionControl.close = Boolean(_inXForwardedFor), new Message({
        status: 403
      }, 'Access denied')
    )
  )
  .pipeline('deny_in')
  .replaceStreamStart(
    new StreamEnd('ConnectionReset')
  )

  // outbound
  .listen(config?.Outbound || config?.Spec?.Traffic?.EnableEgress ? 15001 : 0, {
    'transparent': true,
    'closeEOF': false
    // 'readTimeout': '5s'
  })
  .handleStreamStart(
    (_target) => (
      _outPort = (__inbound?.destinationPort ? __inbound.destinationPort : '0'),
      _outIP = (__inbound?.destinationAddress ? __inbound.destinationAddress : '127.0.0.1'),
      (_outMatch = (_outTrafficMatches && (
        _outTrafficMatches.find(o => ((o.Port == _outPort) && o.DestinationIPRanges_ && o.DestinationIPRanges_.find(e => e.contains(_outIP)))) ||
        _outTrafficMatches.find(o => ((o.Port == _outPort) && !Boolean(o.DestinationIPRanges_) &&
          (o.Protocol == 'http' || o.Protocol == 'https' || (o.Protocol == 'tcp' && o.AllowedEgressTraffic))))))) && (
        ((_outMatch['Protocol'] == 'http') && (_outProtocol = 'http')) ||
        ((_target = (_outMatch['TargetClusters_'] && _outMatch['TargetClusters_'].select())) &&
          _outClustersConfigs[_target] && (_outTarget = _outClustersConfigs[_target]?.select()))
      ),
      (_outProtocol != 'http') && !Boolean(_outTarget) && (
        (_specEnableEgress || (_outMatch && _outMatch.AllowedEgressTraffic)) &&
        (_outTarget = _outIP + ':' + _outPort)
      ),
      _outSessionControl = {
        close: false
      },

      console.log('_outPort: ' + _outPort),
      console.log('_outIP ' + _outIP),
      console.log('_outProtocol: ' + _outProtocol),
      console.log('o_target: ' + _target),
      console.log('_outTarget: ' + _outTarget)

    )
  )
  .link(
    'http_out', () => _outProtocol == 'http',
    'connection_out', () => Boolean(_outTarget),
    'deny_out'
  )
  .pipeline('http_out')
  .demuxHTTP('outbound')
  .replaceMessageStart(
    evt => _outSessionControl.close ? new StreamEnd : evt
  )
  .pipeline('outbound')
  .handleMessageStart(
    (msg, _service, _route, _match, _target) => (
      msg.head.headers?.host && (_service = _outMatch.HttpHostPort2Service[msg.head.headers.host]) &&
      (_route = _outMatch.HttpServiceRouteRules[_service]) &&
      (_match = Object.values(_route).find(o => (
        o.path.exec(msg.head.path) &&
        (!o.Methods_ || o.Methods_[msg.head.method] || o.Methods_['*']) &&
        (!o.AllowedServices_ || o.AllowedServices_[msg.head.headers['serviceidentity']] || o.AllowedServices_['*']) &&
        (!o.Headers_ || Object.entries(o.Headers_).every(
          h => msg.head.headers[h[0]] && h[1].exec(msg.head.headers[h[0]])
        ))))) &&
      (_target = _match.TargetClusters_.select()) && (_outTarget = _outClustersConfigs[_target].select()) &&
      (msg.head.headers['serviceidentity'] = _outMatch.ServiceIdentity),
      !Boolean(_outTarget) && (_specEnableEgress || (_outMatch && _outMatch.AllowedEgressTraffic)) &&
      (_outTarget = _outIP + ':' + _outPort),
      _outTarget && _targetCount.withLabels(_outTarget).increase(),

      console.log(msg.head),
      console.log('o_service: ' + _service),
      console.log('o_route: ' + _route),
      console.log('o_match: ' + _match),
      console.log('o_target: ' + _target),
      console.log('_outTarget: ' + _outTarget)

    )
  )
  .link(
    'request_out', () => Boolean(_outTarget),
    'deny_out_http'
  )
  .pipeline('request_out')
  .muxHTTP(
    'connection_out',
    () => _outTarget
  )
  .pipeline('connection_out')
  .connect(
    () => _outTarget
  )
  .pipeline('deny_out_http')
  .replaceMessage(
    msg => (
      _outSessionControl.close = false, new Message({
        status: 403
      }, 'Access denied')
    )
  )
  .pipeline('deny_out')
  .replaceStreamStart(
    new StreamEnd('ConnectionReset')
  )

  // .listen(14001)
  // .serveHTTP(
  //   new Message('Hi, there!\n')
  // )

  .listen(config?._probeScheme ? 15901 : 0)
  .link(
    'http_liveness', () => config._probeScheme == 'HTTP',
    'connection_liveness', () => Boolean(config._probeTarget),
    'deny_liveness'
  )
  .pipeline('http_liveness')
  .demuxHTTP('message_liveness')
  .pipeline('message_liveness')
  .handleMessageStart(
    msg => (
      msg.head.path == '/osm-liveness-probe' && (msg.head.path = '/liveness'),
      config._probePath && (msg.head.path = config._probePath),
      console.log('probe: ' + config._probeTarget + msg.head.path)
    )
  )
  .muxHTTP('connection_liveness', config?._probeTarget)
  .pipeline('connection_liveness')
  .connect(() => config?._probeTarget)
  .pipeline('deny_liveness')
  .replaceStreamStart(
    new StreamEnd('ConnectionReset')
  )

  .listen(config?._probeScheme ? 15902 : 0)
  .link(
    'http_readiness', () => config._probeScheme == 'HTTP',
    'connection_readiness', () => Boolean(config._probeTarget),
    'deny_readiness'
  )
  .pipeline('http_readiness')
  .demuxHTTP('message_readiness')
  .pipeline('message_readiness')
  .handleMessageStart(
    msg => (
      msg.head.path == '/osm-readiness-probe' && (msg.head.path = '/readiness'),
      config._probePath && (msg.head.path = config._probePath),
      console.log('probe: ' + config._probeTarget + msg.head.path)
    )
  )
  .muxHTTP('connection_readiness', config?._probeTarget)
  .pipeline('connection_readiness')
  .connect(() => config?._probeTarget)
  .pipeline('deny_readiness')
  .replaceStreamStart(
    new StreamEnd('ConnectionReset')
  )

  .listen(config?._probeScheme ? 15903 : 0)
  .link(
    'http_startup', () => config._probeScheme == 'HTTP',
    'connection_startup', () => Boolean(config._probeTarget),
    'deny_startup'
  )
  .pipeline('http_startup')
  .demuxHTTP('message_startup')
  .pipeline('message_startup')
  .handleMessageStart(
    msg => (
      msg.head.path == '/osm-startup-probe' && (msg.head.path = '/startup'),
      config._probePath && (msg.head.path = config._probePath),
      console.log('probe: ' + config._probeTarget + msg.head.path)
    )
  )
  .muxHTTP('connection_startup', config?._probeTarget)
  .pipeline('connection_startup')
  .connect(() => config?._probeTarget)
  .pipeline('deny_startup')
  .replaceStreamStart(
    new StreamEnd('ConnectionReset')
  )

  .listen(15010)
  .link('http_prometheus')
  .pipeline('http_prometheus')
  .demuxHTTP('message_prometheus')
  .pipeline('message_prometheus')
  .handleMessageStart(
    msg => (
      (msg.head.path == '/stats/prometheus' && (msg.head.path = '/metrics')) || (msg.head.path = '/stats' + msg.head.path),
      console.log('prometheus: ' + _prometheusTarget + msg.head.path)
    )
  )
  .muxHTTP('connection_prometheus', () => _prometheusTarget)
  .pipeline('connection_prometheus')
  .connect(() => _prometheusTarget)

  .listen(15000)
  .serveHTTP(
    msg =>
    http.File.from('pipy.json').toMessage(msg.head.headers['accept-encoding'])
  )

)(JSON.decode(pipy.load('pipy.json')))
