package aflag

// TinkerbellBootstrapIP is used to override the Tinkerbell IP for serving a Tinkerbell stack
// from an admin machine.
var TinkerbellBootstrapIP = Flag[string]{
	Name:  "tinkerbell-bootstrap-ip",
	Usage: "The IP used to expose the Tinkerbell stack from the bootstrap cluster",
}

// TinkerbellBMCConsumerURL is a Rufio RPC provider option.
// ConsumerURL is the URL where an rpc consumer/listener is running and to which we will send and receive all notifications.
var TinkerbellBMCConsumerURL = Flag[string]{
	Name:  "tinkerbell-bmc-consumer-url",
	Usage: "The URL of a BMC RPC consumer/listener used for BMC interactions",
}

// TinkerbellBMCHTTPContentType is a Rufio RPC provider option.
// The content type header to use for the rpc request notification.
var TinkerbellBMCHTTPContentType = Flag[string]{
	Name:  "tinkerbell-bmc-http-content-type",
	Usage: "The HTTP content type used for the RPC BMC interactions",
}

// TinkerbellBMCHTTPMethod is a Rufio RPC provider option.
// The HTTP method to use for the rpc request notification.
var TinkerbellBMCHTTPMethod = Flag[string]{
	Name:  "tinkerbell-bmc-http-method",
	Usage: "The HTTP method used for the RPC BMC interactions",
}

// TinkerbellBMCTimestampHeader is a Rufio RPC provider option.
// The the header name that should contain the timestamp.
// Example: X-BMCLIB-Timestamp (in RFC3339 format)
// .
var TinkerbellBMCTimestampHeader = Flag[string]{
	Name:  "tinkerbell-bmc-timestamp-header",
	Usage: "The HTTP timestamp header used for the RPC BMC interactions",
}

// TinkerbellBMCStaticHeaders is a Rufio RPC provider option.
// Predefined headers that will be added to every request (comma separated, values are semicolon separated)
// Example: "X-My-Header=1;2;3,X-Custom-Header=abc;def"
// .
var TinkerbellBMCStaticHeaders = Flag[Header]{
	Name:  "tinkerbell-bmc-static-headers",
	Usage: "Static HTTP headers added to all RPC BMC interactions",
}

// TinkerbellBMCSigHeaderName is a Rufio RPC provider option.
// The header name that should contain the signature(s).
// Example: X-BMCLIB-Signature
// .
var TinkerbellBMCSigHeaderName = Flag[string]{
	Name:  "tinkerbell-bmc-sig-header-name",
	Usage: "The HTTP header name for the HMAC signature used in RPC BMC interactions",
}

// TinkerbellBMCAppendAlgoToHeaderDisabled is a Rufio RPC provider option.
// decides whether to append the algorithm to the signature header or not.
// Example: X-BMCLIB-Signature becomes X-BMCLIB-Signature-256
// .
var TinkerbellBMCAppendAlgoToHeaderDisabled = Flag[bool]{
	Name:  "tinkerbell-bmc-append-algo-to-header-disabled",
	Usage: "This disables appending of the algorithm type to the signature header used in RPC BMC interactions",
}

// TinkerbellBMCSigIncludedPayloadHeaders is a Rufio RPC provider option.
// The headers whose values will be included in the signature payload.
// Example: given these headers in a request: X-My-Header=123,X-Another=456,
// and IncludedPayloadHeaders := []string{"X-Another"}, the value of "X-Another"
// will be included in the signature payload (comma separated).
var TinkerbellBMCSigIncludedPayloadHeaders = Flag[[]string]{
	Name:  "tinkerbell-bmc-sig-included-payload-headers",
	Usage: "The HTTP headers to be included in the HMAC signature payload used in RPC BMC interactions",
}

// TinkerbellBMCPrefixSigDisabled is a Rufio RPC provider option.
// Example: sha256=abc123 ; Determines whether the algorithm will be prefixed to the signature.
var TinkerbellBMCPrefixSigDisabled = Flag[bool]{
	Name:  "tinkerbell-bmc-prefix-sig-disabled",
	Usage: "This disables prefixing the signature with the algorithm type used in RPC BMC interactions",
}

// TinkerbellBMCHMACSecrets is a Rufio RPC provider option.
// secrets used for signing the payload, all secrets with used to sign with both sha256 and sha512.
var TinkerbellBMCHMACSecrets = Flag[[]string]{
	Name:  "tinkerbell-bmc-hmac-secrets",
	Usage: "The secrets used to HMAC sign a payload, used in RPC BMC interactions",
}

// TinkerbellBMCCustomPayload allows providing a JSON payload that will be used in the RPC request.
var TinkerbellBMCCustomPayload = Flag[string]{
	Name:  "tinkerbell-bmc-custom-payload",
	Usage: "The custom payload used in RPC BMC interactions, must be used with tinkerbell-bmc-custom-payload-dot-location",
}

// TinkerbellBMCCustomPayloadDotLocation is the path to where the bmclib RequestPayload{} will be embedded. For example: object.data.body.
var TinkerbellBMCCustomPayloadDotLocation = Flag[string]{
	Name:  "tinkerbell-bmc-custom-payload-dot-location",
	Usage: "The dot location of the custom payload used in RPC BMC interactions, must be used with tinkerbell-bmc-custom-payload",
}
