(
  (
    namespace = (os.env.POD_NAMESPACE || 'default'),
    kind = (os.env.POD_CONTROLLER_KIND || 'Deployment'),
    name = (os.env.SERVICE_ACCOUNT || ''),
    pod = (os.env.POD_NAME || ''),
    address = os.env.REMOTE_LOGGING_ADDRESS,
    logLogging = null,
    logger = {},
  ) => (

    address && (logLogging = new logging.JSONLogger('access-logger').toHTTP('http://' + address +
      (os.env.REMOTE_LOGGING_ENDPOINT || '/?query=insert%20into%20log(message)%20format%20JSONAsString'), {
      batch: {
        prefix: '[',
        postfix: ']',
        separator: ','
      },
      headers: {
        'Content-Type': 'application/json',
        'Authorization': os.env.REMOTE_LOGGING_AUTHORIZATION || ''
      }
    }).log),

    logger.initTracingHeaders = (namespace, kind, name, pod, headers) => (
      (
        uuid = algo.uuid(),
        id = uuid.substring(0, 18).replaceAll('-', ''),
      ) => (
        headers['x-forwarded-proto'] = 'http',
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
      )
    )(),

    logger.loggingEnabled = Boolean(logLogging),

    logger.makeLoggingData = (msg, remoteAddr, remotePort, localAddr, localPort) => (
      msg.head.headers && !msg.head.headers['x-b3-traceid'] && (
        logger.initTracingHeaders(namespace, kind, name, pod, msg.head.headers)
      ),
      {
        reqTime: Date.now(),
        meshName: os.env.MESH_NAME || '',
        remoteAddr,
        remotePort,
        localAddr,
        localPort,
        node: {
          ip: os.env.POD_IP || '127.0.0.1',
          name: os.env.HOSTNAME || 'localhost',
        },
        pod: {
          ns: os.env.POD_NAMESPACE || 'default',
          ip: os.env.POD_IP || '127.0.0.1',
          name: os.env.POD_NAME || os.env.HOSTNAME || 'localhost',
        },
        trace: {
          id: msg.head.headers?.['x-b3-traceid'] || '',
          span: msg.head.headers?.['x-b3-spanid'] || '',
          parent: msg.head.headers?.['x-b3-parentspanid'] || '',
          sampled: msg.head.headers?.['x-b3-sampled'] || ''
        },
        req: Object.assign({ reqSize: msg.body.size, body: msg.body.toString('base64') }, msg.head)
      }
    ),

    logger.saveLoggingData = (loggingData, msg, service, target, ingressEnable, egressEnable, type) => (
      loggingData.service = {
        name: service || 'anonymous', target: target, ingressMode: ingressEnable, egressMode: egressEnable
      },
      loggingData.res = Object.assign({}, msg.head),
      loggingData.res['resSize'] = msg.body.size,
      loggingData.res['body'] = msg.body.toString('base64'),
      loggingData['resTime'] = Date.now(),
      loggingData['endTime'] = Date.now(),
      loggingData['type'] = type,
      logLogging(loggingData)
      // , console.log('loggingData : ', loggingData)
    ),

    logger
  )
)()
