package repo

import (
	"github.com/pkg/errors"
)

var errTooManyConnections = errors.New("too many connections")
var errServiceAccountMismatch = errors.New("service account mismatch in nodeid vs xds certificate common name")
