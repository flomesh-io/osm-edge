(
  (
    {
      debug,
    } = pipy.solve('utils.js'),
    tracing
  ) => (

    tracing = {
      logZipkin: null,
      tracingAddress: os.env.TRACING_ADDRESS,
      tracingEndpoint: (os.env.TRACING_ENDPOINT || '/api/v2/spans'),
    },

    tracing.tracingAddress &&
    (tracing.logZipkin = new logging.JSONLogger('zipkin').toHTTP('http://' + tracing.tracingAddress + tracing.tracingEndpoint, {
      batch: {
        prefix: '[',
        postfix: ']',
        separator: ','
      },
      headers: {
        'Host': tracing.tracingAddress,
        'Content-Type': 'application/json',
      }
    }).log),

    tracing.initTracingHeaders = (namespace, kind, name, pod, headers, proto, uuid, id) => (
      uuid = algo.uuid(),
      id = uuid.substring(0, 18).replaceAll('-', ''),
      proto && (headers['x-forwarded-proto'] = proto),
      headers['x-b3-spanid'] &&
      (headers['x-b3-parentspanid'] = headers['x-b3-spanid']) &&
      (headers['x-b3-spanid'] = id),
      !headers['x-b3-traceid'] &&
      (headers['x-b3-traceid'] = id) &&
      (headers['x-b3-spanid'] = id) &&
      (headers['x-b3-sampled'] = '1'),
      !headers['x-request-id'] && (headers['x-request-id'] = uuid),
      headers['osm-stats-namespace'] = namespace,
      headers['osm-stats-kind'] = kind,
      headers['osm-stats-name'] = name,
      headers['osm-stats-pod'] = pod
    ),

    tracing.makeZipKinData = (name, msg, headers, clusterName, kind, shared, data) => (
      data = {
        'traceId': headers?.['x-b3-traceid'] && headers['x-b3-traceid'].toString(),
        'id': headers?.['x-b3-spanid'] && headers['x-b3-spanid'].toString(),
        'name': headers?.host,
        'timestamp': Date.now() * 1000,
        'localEndpoint': {
          'port': 0,
          'ipv4': os.env.POD_IP || '',
          'serviceName': name,
        },
        'tags': {
          'component': 'proxy',
          'http.url': headers?.['x-forwarded-proto'] + '://' + headers?.host + msg?.head?.path,
          'http.method': msg?.head?.method,
          'node_id': os.env.POD_UID || '',
          'http.protocol': msg?.head?.protocol,
          'guid:x-request-id': headers?.['x-request-id'],
          'user_agent': headers?.['user-agent'],
          'upstream_cluster': clusterName
        },
        'annotations': []
      },
      headers['x-b3-parentspanid'] && (data['parentId'] = headers['x-b3-parentspanid']),
      data['kind'] = kind,
      shared && (data['shared'] = shared),
      data.tags['request_size'] = '0',
      data.tags['response_size'] = '0',
      data.tags['http.status_code'] = '502',
      data.tags['peer.address'] = '',
      data['duration'] = 0,
      data
    ),

    tracing.save = (zipkinData, messageHead, bytesStruct, target) => (
      zipkinData && (
        zipkinData.tags['peer.address'] = target,
        zipkinData.tags['http.status_code'] = messageHead.status?.toString?.(),
        zipkinData.tags['request_size'] = bytesStruct.requestSize.toString(),
        zipkinData.tags['response_size'] = bytesStruct.responseSize.toString(),
        zipkinData['duration'] = Date.now() * 1000 - zipkinData['timestamp'],
        tracing.logZipkin(zipkinData),
        debug(log => log('zipkinData : ', zipkinData))
      )
    ),

    tracing
  )
)()