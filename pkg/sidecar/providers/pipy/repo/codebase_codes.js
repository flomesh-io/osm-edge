// version: '2022.07.18'
(code) => (
  ((
    codes = {

      "NoService": {
        'ConfigBlock': 'HttpHostPort2Service',
        'ErrorCode': 'no_service',
        'ErrorDescription': 'Can\'t find service by `HTTP host`'
      },

      "NoRoute": {
        'ConfigBlock': 'HttpServiceRouteRules',
        'ErrorCode': 'no_route',
        'ErrorDescription': 'Can\'t find route by `service`'
      },

      "NoEndpoint": {
        'ConfigBlock': 'ClustersConfigs',
        'ErrorCode': 'no_endpoint',
        'ErrorDescription': 'Can\'t find endpoint by `route`'
      }

    },

    msg = codes[code]
  ) => (
    msg ? '[' + msg.ConfigBlock + '] ' + msg.ErrorCode + ' (' + msg.ErrorDescription + ')' : 'Invalid error code!'
  ))()
)