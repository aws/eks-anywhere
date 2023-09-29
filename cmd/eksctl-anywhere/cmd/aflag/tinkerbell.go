package aflag

// TinkerbellBootstrapIP is used to override the Tinkerbell IP for serving a Tinkerbell stack
// from an admin machine.
var TinkerbellBootstrapIP = Flag[string]{
	Name:  "tinkerbell-bootstrap-ip",
	Usage: "The IP used to expose the Tinkerbell stack from the bootstrap cluster",
}

// TinkerbellBMCConsumerURL is a Rufio RPC provider option.
var TinkerbellBMCConsumerURL = Flag[string]{
	Name:  "tinkerbell-bmc-consumer-url",
	Usage: "The URL used to expose the Tinkerbell BMC consumer from the bootstrap cluster",
}

// TinkerbellBMCHTTPContentType is a Rufio RPC provider option.
var TinkerbellBMCHTTPContentType = Flag[string]{
	Name:  "tinkerbell-bmc-http-content-type",
	Usage: "The HTTP content type used to expose the Tinkerbell BMC consumer from the bootstrap cluster",
}

// TinkerbellBMCHTTPMethod is a Rufio RPC provider option.
var TinkerbellBMCHTTPMethod = Flag[string]{
	Name:  "tinkerbell-bmc-http-method",
	Usage: "The HTTP method used to expose the Tinkerbell BMC consumer from the bootstrap cluster",
}

// TinkerbellBMCTimestampHeader is a Rufio RPC provider option.
var TinkerbellBMCTimestampHeader = Flag[string]{
	Name:  "tinkerbell-bmc-timestamp-header",
	Usage: "The HTTP timestamp header used to expose the Tinkerbell BMC consumer from the bootstrap cluster",
}

// TinkerbellBMCStaticHeaders is a Rufio RPC provider option.
var TinkerbellBMCStaticHeaders = Flag[string]{
	Name:  "tinkerbell-bmc-static-headers",
	Usage: "The HTTP static headers used to expose the Tinkerbell BMC consumer from the bootstrap cluster",
}

// TinkerbellBMCHeaderName is a Rufio RPC provider option.
var TinkerbellBMCHeaderName = Flag[string]{
	Name:  "tinkerbell-bmc-header-name",
	Usage: "The HTTP header name used to expose the Tinkerbell BMC consumer from the bootstrap cluster",
}

// TinkerbellBMCAppendAlgoToHeaderDisabled is a Rufio RPC provider option.
var TinkerbellBMCAppendAlgoToHeaderDisabled = Flag[bool]{
	Name:  "tinkerbell-bmc-append-algo-to-header-disabled",
	Usage: "The HTTP append algo to header disabled used to expose the Tinkerbell BMC consumer from the bootstrap cluster",
}

// TinkerbellBMCIncludedPayloadHeaders is a Rufio RPC provider option.
var TinkerbellBMCIncludedPayloadHeaders = Flag[[]string]{
	Name:  "tinkerbell-bmc-included-payload-headers",
	Usage: "The HTTP included payload headers used to expose the Tinkerbell BMC consumer from the bootstrap cluster. If you specify a Timestamp header, it must be included here.",
}

// TinkerbellBMCPrefixSigDisabled is a Rufio RPC provider option.
var TinkerbellBMCPrefixSigDisabled = Flag[bool]{
	Name:  "tinkerbell-bmc-prefix-sig-disabled",
	Usage: "The HTTP prefix sig disabled used to expose the Tinkerbell BMC consumer from the bootstrap cluster",
}

// TinkerbellBMCWebhookSecrets is a Rufio RPC provider option.
var TinkerbellBMCWebhookSecrets = Flag[[]string]{
	Name:  "tinkerbell-bmc-webhook-secrets",
	Usage: "The webhook secrets used to expose the Tinkerbell BMC consumer from the bootstrap cluster",
}
