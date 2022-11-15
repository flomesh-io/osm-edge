((
  logger = pipy.solve('logging-init.js')
) => (

  pipy({
    target: null,
    service: null,
    loggingData: null
  })

    .import({
      _flow: 'main',
      _inTarget: 'inbound-classifier',
      _ingressEnable: 'inbound-classifier',
      _inService: 'inbound-http-routing',
      _outTarget: 'outbound-classifier',
      _egressEnable: 'outbound-classifier',
      _outService: 'outbound-http-routing'
    })

    .pipeline()

    .onStart(
      () => (
        logger.logLogging && void (
          (_flow === 'inbound') && (
            target = _inTarget?.id,
            service = _inService
          ),
          (_flow === 'outbound') && (
            target = _outTarget.id,
            service = _outService
          )
        )
      )
    )

    .handleMessage(
      (msg) => (
        logger.logLogging && (
          loggingData = {
            reqTime: Date.now(),
            meshName: os.env.MESH_NAME || '',
            remoteAddr: __inbound?.remoteAddress,
            remotePort: __inbound?.remotePort,
            localAddr: __inbound?.destinationAddress,
            localPort: __inbound?.destinationPort,
            node: {
              ip: os.env.POD_IP || '127.0.0.1',
              name: os.env.HOSTNAME || 'localhost',
            },
            pod: {
              ns: os.env.POD_NAMESPACE || 'default',
              ip: os.env.POD_IP || '127.0.0.1',
              name: os.env.POD_NAME || os.env.HOSTNAME || 'localhost',
            },
            service: {
              name: service || 'anonymous', target: target, ingressMode: _ingressEnable, egressMode: _egressEnable
            },
            trace: {
              id: msg.head.headers?.['x-b3-traceid'] || '',
              span: msg.head.headers?.['x-b3-spanid'] || '',
              parent: msg.head.headers?.['x-b3-parentspanid'] || '',
              sampled: msg.head.headers?.['x-b3-sampled'] || ''
            }
          },
          loggingData.req = Object.assign({}, msg.head),
          loggingData.req['reqSize'] = msg.body.size,
          loggingData.req['body'] = msg.body.toString('base64')
        )
      )
    )

    .chain()

    .handleMessage(
      msg => (
        logger.logLogging && (
          loggingData.res = Object.assign({}, msg.head),
          loggingData.res['resSize'] = msg.body.size,
          loggingData.res['body'] = msg.body.toString('base64'),
          loggingData['resTime'] = Date.now(),
          loggingData['endTime'] = Date.now(),
          loggingData['type'] = _flow,
          logger.save(loggingData)
        )
      )
    )

))()