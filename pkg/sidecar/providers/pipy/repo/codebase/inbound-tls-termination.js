((
  {
    tlsCertChain,
    tlsPrivateKey,
    tlsIssuingCA,
  } = pipy.solve('config.js'),
  {
    debug,
  } = pipy.solve('utils.js'),

  sourceIPRangesCache = new algo.Cache((sourceIPRanges) => (
    sourceIPRanges ? Object.entries(sourceIPRanges).map(([k, v]) =>
    ({
      netmask: new Netmask(k),
      mTLS: v?.mTLS,
      skipClientCertValidation: v?.SkipClientCertValidation,
      authenticatedPrincipals: v?.AuthenticatedPrincipals && Object.fromEntries(v.AuthenticatedPrincipals.map(e => [e, true])),
    })) : null
  ), null, {})

) => (

  pipy({
    _tlsStruct: null,
    _forbiddenTLS: false,
  })

    .import({
      _inMatch: 'inbound-classifier',
      _ingressEnable: 'inbound-classifier'
    })

    //
    // accept tls connection
    //
    .pipeline()

    .onStart(
      () => (
        ((remoteAddress = __inbound.remoteAddress || '127.0.0.1', sourceIPRanges) => (

          _inMatch?.SourceIPRanges && (sourceIPRanges = sourceIPRangesCache.get(_inMatch?.SourceIPRanges)) && (
            // INGRESS mode
            _tlsStruct = sourceIPRanges?.find?.(e => e?.netmask?.contains?.(remoteAddress)),
            _ingressEnable = Boolean(_tlsStruct)
          ),

          debug(log => log('inbound acceptTLS - TLS/_ingressEnable: ', Boolean(tlsCertChain), _ingressEnable)),
          null
        ))()
      )
    )

    .branch(
      () => Boolean(tlsCertChain) && Boolean(_inMatch) && (!Boolean(_tlsStruct) || _tlsStruct?.mTLS), $ => $
        .acceptTLS({
          certificate: () => ({
            cert: new crypto.Certificate(tlsCertChain),
            key: new crypto.PrivateKey(tlsPrivateKey),
          }),
          trusted: (!tlsIssuingCA && []) || [
            new crypto.Certificate(tlsIssuingCA),
          ],
          verify: (ok, cert) => (
            _tlsStruct?.mTLS && !Boolean(_tlsStruct?.skipClientCertValidation) && (
              _tlsStruct?.authenticatedPrincipals && (_forbiddenTLS = true),
              (_tlsStruct?.authenticatedPrincipals?.[cert?.subject?.commonName] ||
                (cert?.subjectAltNames && cert.subjectAltNames.find(o => _tlsStruct?.authenticatedPrincipals?.[o]))) && (
                _forbiddenTLS = false
              ),
              _forbiddenTLS && (
                (_inMatch.Protocol !== 'http' && _inMatch.Protocol !== 'grpc') && (
                  ok = false
                ),
                debug(log => log('Bad client certificate :', cert?.subject))
              )
            ),
            ok
          )
        }).to($ => $
          .branch(
            () => _forbiddenTLS, $ => $
              .demuxHTTP().to($ => $
                .replaceMessage(
                  new Message({
                    status: 403
                  }, 'Access denied')
                )
              ),

            $ => $
              .chain()
          )
        ),

      $ => $
        .chain()
    )

))()