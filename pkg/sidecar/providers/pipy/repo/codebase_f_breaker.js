// version: '2022.07.28a'
(
  serviceName = '',
  maxConnections = 10,
  maxRequestsPerConnection = 10,
  maxPendingRequests = 10,
  statTimeWindow = 15, // 15s
  slowTimeThreshold = 3, // 3s
  slowAmountThreshold = 10,
  slowRatioThreshold = 0.5,
  errorAmountThreshold = 10,
  errorRatioThreshold = 0.25,
  degradedTimeWindow = 15, // 15s
  degradedStatusCode = 409,
  degradedResponseContent = 'Coming soon ...'
) => (
  ((
    total = 0,
    slowAmount = 0,
    errorAmount = 0,
    degraded = false,
    lastDegraded = false,
    tcpConnections = 0,
    slowQuota = new algo.Quota(slowAmountThreshold, {
      per: statTimeWindow
    }),
    errorQuota = new algo.Quota(errorAmountThreshold, {
      per: statTimeWindow
    }),
    degradedQuota = new algo.Quota(1, {
      per: degradedTimeWindow
    }),
    samplingSlowQuota = new algo.Quota((statTimeWindow + 4) / 5 - 1, {
      per: statTimeWindow
    }),
    samplingErrorQuota = new algo.Quota((statTimeWindow + 4) / 5 - 1, {
      per: statTimeWindow
    }),
    report,
    exceedMaxConnections
  ) => (

    console.log('serviceName:', serviceName),
    console.log('maxConnections:', maxConnections),
    console.log('maxRequestsPerConnection:', maxRequestsPerConnection),
    console.log('maxPendingRequests:', maxPendingRequests),
    console.log('statTimeWindow:', statTimeWindow),
    console.log('slowTimeThreshold:', slowTimeThreshold),
    console.log('slowAmountThreshold:', slowAmountThreshold),
    console.log('slowRatioThreshold:', slowRatioThreshold),
    console.log('errorAmountThreshold:', errorAmountThreshold),
    console.log('errorRatioThreshold:', errorRatioThreshold),
    console.log('degradedTimeWindow:', degradedTimeWindow),
    console.log('degradedStatusCode:', degradedStatusCode),
    console.log('degradedResponseContent:', degradedResponseContent),

    report = code => (
      lastDegraded = degraded,
      ((code & 0x1) == 1) && (++slowAmount) && (slowQuota.consume(1) != 1) && (degraded = true),
      ((code & 0x2) == 2) && (++errorAmount) && (errorQuota.consume(1) != 1) && (degraded = true),
      !lastDegraded && degraded && degradedQuota.consume(1)
    ),

    exceedMaxConnections = () => (
      tcpConnections >= maxConnections ? console.log('=== [circuit_breaker] === (exceedMaxConnections)', serviceName, tcpConnections, maxConnections) || true : false
    ),

    {
      increase: () => (
        ++total
      ),

      block: () => (
        degraded && (degradedQuota.consume(1) == 1) && (lastDegraded = degraded = false),
        degraded && console.log('=== [circuit_breaker] === (block)', serviceName, degraded),
        degraded || exceedMaxConnections()
      ),

      checkSlow: seconds => (
        (seconds >= slowTimeThreshold) && report(0x1)
      ),

      checkStatusCode: statusCode => (
        (statusCode < 200 || statusCode > 299) && report(0x2)
      ),

      sample: () => (
        (total > 0) && (() => (
          lastDegraded = degraded,
          (slowAmount / total >= slowRatioThreshold) && (samplingSlowQuota.consume(1) != 1) && (degraded = true),
          (errorAmount / total >= errorRatioThreshold) && (samplingErrorQuota.consume(1) != 1) && (degraded = true),
          !lastDegraded && degraded && degradedQuota.consume(1),
          total = slowAmount = errorAmount = 0,
          degraded && console.log('=== [circuit_breaker] === (timer)', serviceName, degraded)
        ))()
      ),

      message: () => (
        new Message({
          status: degradedStatusCode
        }, degradedResponseContent)
      ),

      incConnections: () => (
        ++tcpConnections
      ),

      decConnections: () => (
        --tcpConnections
      ),

      serviceName: () => serviceName,

      maxConnections: () => maxConnections,

      maxRequestsPerConnection: () => maxRequestsPerConnection,

      maxPendingRequests: () => maxPendingRequests
    }))()
)