package crypto

import (
	"strings"
)

// This is what we currently support as the default. In the future,
// we can make this customizable and return a wider range of
// supported names.
func SecureCipherSuiteNames() []string {
	return []string{"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"}
}

func SecureCipherSuitesString() string {
	return strings.Join(SecureCipherSuiteNames(), ",")
}
