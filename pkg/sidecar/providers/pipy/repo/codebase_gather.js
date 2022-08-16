// version: '2022.08.12'
((
  {
    debugLogLevel,
    namespace,
    kind,
    name,
    pod,
    metrics
  } = pipy.solve('config.js')) => (

  pipy({
    _inMessageHead: null,
    _outMessageHead: null
  })

    .import({
      logZipkin: 'main',
      logLogging: 'main',
      _inTarget: 'main',
      _inZipkinData: 'main',
      _inLoggingData: 'main',
      _inBytesStruct: 'main',
      _localClusterName: 'main',
      _outTarget: 'main',
      _outBytesStruct: 'main',
      _outRequestTime: 'main',
      _upstreamClusterName: 'main',
      _outZipkinData: 'main',
      _outLoggingData: 'main'
    })

    //
    // Metrics and Logging after local service response. 
    //
    .pipeline('after-local-http')
    .handleMessageStart(
      (msg) => (
        _inMessageHead = msg?.head,
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
      )
    )
    .handleMessageEnd(
      () => (
        _inMessageHead && (() => (
          _inZipkinData && (() => (
            _inZipkinData.tags['peer.address'] = _inTarget?.id,
            _inZipkinData.tags['http.status_code'] = _inMessageHead.status?.toString?.(),
            _inZipkinData.tags['request_size'] = _inBytesStruct.requestSize.toString(),
            _inZipkinData.tags['response_size'] = _inBytesStruct.responseSize.toString(),
            _inZipkinData['duration'] = Date.now() * 1000 - _inZipkinData['timestamp'],
            debugLogLevel && console.log('_inZipkinData : ', _inZipkinData),
            logZipkin(_inZipkinData)
          ))(),
          _inLoggingData && (() => (
            _inLoggingData['downstream'] = __inbound.remoteAddress,
            _inLoggingData['upstream'] = _inTarget?.id,
            _inLoggingData['response_code'] = _inMessageHead.status,
            _inLoggingData['duration'] = Date.now() - _inLoggingData['start_time'],
            _inLoggingData['start_time'] = _inLoggingData['start_time'].toISOString(),
            _inLoggingData['bytes_received'] = _inBytesStruct.requestSize,
            _inLoggingData['bytes_sent'] = _inBytesStruct.responseSize,
            logLogging(_inLoggingData)
          ))()
        ))()
      )
    )

    //
    // Metrics and Logging after upstream service response. 
    //
    .pipeline('after-upstream-http')
    .handleMessageStart(
      (msg) => (
        _outMessageHead = msg?.head,
        ((headers, d_namespace, d_kind, d_name, d_pod) => (
          headers = msg?.head?.headers,
          (d_namespace = headers?.['osm-stats-namespace']) && (delete headers['osm-stats-namespace']),
          (d_kind = headers?.['osm-stats-kind']) && (delete headers['osm-stats-kind']),
          (d_name = headers?.['osm-stats-name']) && (delete headers['osm-stats-name']),
          (d_pod = headers?.['osm-stats-pod']) && (delete headers['osm-stats-pod']),
          d_namespace && metrics.osmRequestDurationHist.withLabels(namespace, kind, name, pod, d_namespace, d_kind, d_name, d_pod).observe(Date.now() - _outRequestTime),
          metrics.upstreamCompletedCount.withLabels(_upstreamClusterName).increase(),
          msg?.head?.status && metrics.upstreamCodeCount.withLabels(msg.head.status, _upstreamClusterName).increase(),
          msg?.head?.status && metrics.upstreamCodeXCount.withLabels(msg.head.status.toString().charAt(0), _upstreamClusterName).increase(),
          metrics.upstreamResponseTotal.withLabels(namespace, kind, name, pod, _upstreamClusterName).increase(),
          msg?.head?.status && metrics.upstreamResponseCode.withLabels(msg.head.status.toString().charAt(0), namespace, kind, name, pod, _upstreamClusterName).increase()
        ))()
      )
    )
    .handleMessageEnd(
      () => (
        _outMessageHead &&
        (() => (
          _outZipkinData && (() => (
            _outZipkinData.tags['peer.address'] = _outTarget?.id,
            _outZipkinData.tags['http.status_code'] = _outMessageHead.status?.toString?.(),
            _outZipkinData.tags['request_size'] = _outBytesStruct.requestSize.toString(),
            _outZipkinData.tags['response_size'] = _outBytesStruct.responseSize.toString(),
            _outZipkinData['duration'] = Date.now() * 1000 - _outZipkinData['timestamp'],
            debugLogLevel && console.log('_outZipkinData : ', _outZipkinData),
            logZipkin(_outZipkinData)
          ))(),
          _outLoggingData && (() => (
            _outLoggingData['downstream'] = __inbound.remoteAddress,
            _outLoggingData['upstream'] = _outTarget?.id,
            _outLoggingData['response_code'] = _outMessageHead.status,
            _outLoggingData['duration'] = Date.now() - _outLoggingData['start_time'],
            _outLoggingData['start_time'] = _outLoggingData['start_time'].toISOString(),
            _outLoggingData['bytes_received'] = _outBytesStruct.requestSize,
            _outLoggingData['bytes_sent'] = _outBytesStruct.responseSize,
            logLogging(_outLoggingData)
          ))()
        ))()
      )
    )
))()