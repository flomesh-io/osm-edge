package certificate

import (
	"fmt"
	"time"
)

// IssueOption is an option that can be passed to IssueCertificate.
type IssueOption func(*issueOptions)

type issueOptions struct {
	fullCNProvided   bool
	validityDuration *time.Duration
}

func (o *issueOptions) formatCN(prefix, trustDomain string) CommonName {
	if o.fullCNProvided {
		return CommonName(prefix)
	}
	return CommonName(fmt.Sprintf("%s.%s", prefix, trustDomain))
}

func (o *issueOptions) validityPeriod(validityDuration time.Duration) time.Duration {
	if o.validityDuration != nil {
		return *o.validityDuration
	}
	return validityDuration
}

// FullCNProvided tells IssueCertificate that the provided prefix is actually the full trust domain, and not to append
// the issuer's trust domain.
func FullCNProvided() IssueOption {
	return func(opts *issueOptions) {
		opts.fullCNProvided = true
	}
}

// ValidityDurationProvided tells IssueCertificate that the certificate's validity duration.
func ValidityDurationProvided(validityDuration *time.Duration) IssueOption {
	return func(opts *issueOptions) {
		opts.validityDuration = validityDuration
	}
}
