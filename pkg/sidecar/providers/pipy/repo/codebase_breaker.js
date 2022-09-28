// version: '2022.09.28'
(
  serviceName = '',
  maxConnections = 10,
  maxRequestsPerConnection = 10, // unimplement
  maxPendingRequests = 10, // unimplement
  minRequestAmount = 100,
  statTimeWindow = 30, // 30s
  slowTimeThreshold = 5, // 5s
  slowAmountThreshold = 0,
  slowRatioThreshold = 0.0,
  errorAmountThreshold = 0,
  errorRatioThreshold = 0.0,
  degradedTimeWindow = 30, // 30s
  degradedStatusCode = 409,
  degradedResponseContent = 'Coming soon ...'
) => (
  ((
    tick = 0,
    total = 0,
    slowAmount = 0,
    errorAmount = 0,
    degraded = false,
    lastDegraded = false,
    tcpConnections = 0,
    slowQuota = slowAmountThreshold > 0 ? new algo.Quota(slowAmountThreshold - 1, {
      per: statTimeWindow
    }) : null,
    errorQuota = errorAmountThreshold > 0 ? new algo.Quota(errorAmountThreshold - 1, {
      per: statTimeWindow
    }) : null,
    degradedQuota = new algo.Quota(1, {
      per: degradedTimeWindow
    }),
    checkSample,
    report,
    exceedMaxConnections
  ) => (

    console.log('serviceName:', serviceName),
    console.log('maxConnections:', maxConnections),
    console.log('maxRequestsPerConnection:', maxRequestsPerConnection),
    console.log('maxPendingRequests:', maxPendingRequests),
    console.log('minRequestAmount:', minRequestAmount),
    console.log('statTimeWindow:', statTimeWindow),
    console.log('slowTimeThreshold:', slowTimeThreshold),
    console.log('slowAmountThreshold:', slowAmountThreshold),
    console.log('slowRatioThreshold:', slowRatioThreshold),
    console.log('errorAmountThreshold:', errorAmountThreshold),
    console.log('errorRatioThreshold:', errorRatioThreshold),
    console.log('degradedTimeWindow:', degradedTimeWindow),
    console.log('degradedStatusCode:', degradedStatusCode),
    console.log('degradedResponseContent:', degradedResponseContent),

    checkSample = (cond) => (
      !degraded && (total >= minRequestAmount) && (
        lastDegraded = degraded,
        (slowRatioThreshold > 0) && (
          (slowAmount / total >= slowRatioThreshold) && (degraded = true)
        ),
        (errorRatioThreshold > 0) && (
          (errorAmount / total >= errorRatioThreshold) && (degraded = true)
        ),
        !lastDegraded && degraded && degradedQuota.consume(1),
        degraded && console.log('[circuit_breaker] total/slowAmount/errorAmount', cond, serviceName, degraded, total, slowAmount, errorAmount)
      )
    ),

    report = code => (
      lastDegraded = degraded,
      ((code & 0x1) == 1) && (++slowAmount) && slowQuota && (slowQuota.consume(1) != 1) && (degraded = true),
      ((code & 0x2) == 2) && (++errorAmount) && errorQuota && (errorQuota.consume(1) != 1) && (degraded = true),
      !lastDegraded && degraded && degradedQuota.consume(1)
    ),

    {
      increase: () => (
        ++total
      ),

      block: () => (
        checkSample('block'),
        degraded && (degradedQuota.consume(1) == 1) && (lastDegraded = degraded = false),
        // degraded && console.log('[circuit_breaker] (block)', serviceName, degraded),
        degraded
      ),

      checkSlow: seconds => (
        (seconds >= slowTimeThreshold) && report(0x1)
      ),

      checkStatusCode: statusCode => (
        (statusCode < 200 || statusCode > 299) && report(0x2)
      ),

      sample: () => (
        degraded && (
          tick = 0,
          (total > 0) && (
            total = slowAmount = errorAmount = 0
          )
        ),

        !degraded && (
          checkSample('timer'),
          (++tick > statTimeWindow) && (
            tick = total = slowAmount = errorAmount = 0
          )
        )
      ),

      message: () => (
        [
          new Message({ status: degradedStatusCode }, degradedResponseContent),
          new StreamEnd
        ]
      ),

      incConnections: () => (
        ++tcpConnections
      ),

      decConnections: () => (
        --tcpConnections
      ),

      exceedMaxConnections: () => (
        tcpConnections > maxConnections ? console.log('=== [circuit_breaker] === (exceedMaxConnections)', serviceName, tcpConnections, maxConnections) || true : false
      ),

      serviceName: () => serviceName,

      maxConnections: () => maxConnections,

      maxRequestsPerConnection: () => maxRequestsPerConnection,

      maxPendingRequests: () => maxPendingRequests,

      minRequestAmount: () => minRequestAmount

    }))()
)
