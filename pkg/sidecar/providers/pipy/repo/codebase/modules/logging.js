(
  (
    logLogging = null,
    address = os.env.REMOTE_LOGGING_ADDRESS,
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

    logger.loggingEnabled = Boolean(logLogging),

    logger.makeLoggingData = (msg) => (
      {
        reqTime: Date.now(),
        meshName: os.env.MESH_NAME || '',
        remoteAddr: () => __inbound?.remoteAddress,
        remotePort: () => __inbound?.remotePort,
        localAddr: () => __inbound?.destinationAddress,
        localPort: () => __inbound?.destinationPort,
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

    logger.padLoggingData = (loggingData, msg, service, target, ingressEnable, egressEnable, type) => (
      loggingData.service = {
        name: service || 'anonymous', target: target, ingressMode: ingressEnable, egressMode: egressEnable
      },
      loggingData.res = Object.assign({}, msg.head),
      loggingData.res['resSize'] = msg.body.size,
      loggingData.res['body'] = msg.body.toString('base64'),
      loggingData['resTime'] = Date.now(),
      loggingData['endTime'] = Date.now(),
      loggingData['type'] = type
    ),

    logger.saveLogging = (loggingData) => (
      logLogging(loggingData)
      // , console.log('loggingData : ', loggingData)
    ),

    logger
  )
)()
