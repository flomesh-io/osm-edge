(
  (
    config = pipy.solve('config.js'),

    plugins = {
      inboundL7Chains: [
        { 'INBOUND_HTTP_AFTER_TLS': ['modules/inbound-tls-termination.js'] },
        {
          'INBOUND_HTTP_AFTER_ROUTING': ['modules/inbound-http-routing.js',
            'modules/inbound-metrics-http.js',
            'modules/inbound-tracing-http.js',
            'modules/inbound-logging-http.js',
            'modules/inbound-throttle-service.js',
            'modules/inbound-throttle-route.js',
          ]
        },
        { 'INBOUND_HTTP_AFTER_BALANCING': ['modules/inbound-http-load-balancing.js'] },
        { 'INBOUND_HTTP_PLUGINS': [] },
        { 'INBOUND_HTTP_UPSTREAM': ['modules/inbound-upstream.js'] },
        { 'INBOUND_HTTP_DEFAULT': ['modules/inbound-http-default.js',] }
      ],
      inboundL4Chains: [
        { 'INBOUND_TCP_AFTER_TLS': ['modules/inbound-tls-termination.js'] },
        { 'INBOUND_TCP_AFTER_BALANCING': ['modules/inbound-tcp-load-balancing.js'] },
        { 'INBOUND_TCP_PLUGINS': [] },
        { 'INBOUND_TCP_UPSTREAM': ['modules/inbound-upstream.js'] },
        { 'INBOUND_TCP_DEFAULT': ['modules/inbound-tcp-default.js',] }
      ],

      outboundL7Chains: [
        {
          'OUTBOUND_HTTP_AFTER_ROUTING': ['modules/outbound-http-routing.js',
            'modules/outbound-metrics-http.js',
            'modules/outbound-tracing-http.js',
            'modules/outbound-logging-http.js',
            'modules/outbound-circuit-breaker.js',
          ]
        },
        { 'OUTBOUND_HTTP_AFTER_BALANCING': ['modules/outbound-http-load-balancing.js',] },
        { 'OUTBOUND_HTTP_PLUGINS': [] },
        { 'OUTBOUND_HTTP_TLS_INITIATION': ['modules/outbound-tls-initiation.js',] },
        { 'OUTBOUND_HTTP_DEFAULT': ['modules/outbound-http-default.js',] }
      ],
      outboundL4Chains: [
        { 'OUTBOUND_TCP_AFTER_BALANCING': ['modules/outbound-tcp-load-balancing.js',] },
        { 'OUTBOUND_TCP_PLUGINS': [] },
        { 'OUTBOUND_TCP_TLS_INITIATION': ['modules/outbound-tls-initiation.js',] },
        { 'OUTBOUND_TCP_DEFAULT': ['modules/outbound-tcp-default.js',] }
      ]
    },

    findChain = name => (
      ((obj = null) => (
        Object.entries(plugins).map(
          ([k, v]) => (v.forEach(o => o?.[name] && (obj = o?.[name])))
        ),
        obj
      ))()
    ),
    
    expandChains = chains => (
      ((array = []) => (
        chains.map(
          o => (
            Object.entries(o).map(
              ([k, v]) => (
                v.map(
                  c => array.push(c)
                )
              )
            )
          )
        ),
        console.log('Chains:', array.map(o => '\n                              ->[' + o + ']').join('\n').replaceAll('\n\n', '\n')),
        array
      ))()
    )

  ) => (

    (config?.Chains?.['inbound-http'] || []).forEach(
      p => findChain('INBOUND_HTTP_PLUGINS')?.push(p) 
    ),

    (config?.Chains?.['inbound-tcp'] || []).forEach(
      p => findChain('INBOUND_TCP_PLUGINS')?.push(p) 
    ),

    (config?.Chains?.['outbound-http'] || []).forEach(
      p => findChain('OUTBOUND_HTTP_PLUGINS')?.push(p)  
    ),

    (config?.Chains?.['outbound-tcp'] || []).forEach(
      p => findChain('OUTBOUND_TCP_PLUGINS')?.push(p)  
    ),

    {
      inboundL7Chains: expandChains(plugins.inboundL7Chains),
      inboundL4Chains: expandChains(plugins.inboundL4Chains),
      outboundL7Chains: expandChains(plugins.outboundL7Chains),
      outboundL4Chains: expandChains(plugins.outboundL4Chains),
    }

  )
  
)()