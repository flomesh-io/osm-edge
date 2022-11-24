(
  (
    {
      debug,
    } = pipy.solve('utils.js'),
    logger
  ) => (

    logger = {
      address: null,
      logLogging: null,
    },

    logger.address = os.env.REMOTE_LOGGING_ADDRESS,

    logger.address && (logger.logLogging = new logging.JSONLogger('access-logger').toHTTP('http://' + logger.address +
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

    logger.save = (loggingData) => (
      logger.logLogging(loggingData),
      debug(log => log('loggingData : ', loggingData))
    ),

    logger
  )

)()