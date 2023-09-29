package rufio

import (
	"net/http"

	corev1 "k8s.io/api/core/v1"
)

// RedfishOptions contains the redfish provider specific options.
type RedfishOptions struct {
	// Port that redfish will use for calls.
	Port int `json:"port"`
}

// IPMITOOLOptions contains the ipmitool provider specific options.
type IPMITOOLOptions struct {
	// Port that ipmitool will use for calls.
	// +optional
	Port int `json:"port"`
	// CipherSuite that ipmitool will use for calls.
	// +optional
	CipherSuite string `json:"cipherSuite"`
}

// IntelAMTOptions contains the intelAMT provider specific options.
type IntelAMTOptions struct {
	// Port that intelAMT will use for calls.
	Port int `json:"port"`
}

// HMACAlgorithm is a type for HMAC algorithms.
type HMACAlgorithm string

// HMACSecrets holds per Algorithm slice secrets.
// These secrets will be used to create HMAC signatures.
type HMACSecrets map[HMACAlgorithm][]corev1.SecretReference

// RPCOptions defines the configurable options to use when sending rpc notifications.
type RPCOptions struct {
	// ConsumerURL is the URL where an rpc consumer/listener is running
	// and to which we will send and receive all notifications.
	ConsumerURL string `json:"consumerURL"`
	// LogNotificationsDisabled determines whether responses from rpc consumer/listeners will be logged or not.
	// +optional
	LogNotificationsDisabled bool `json:"logNotificationsDisabled"`
	// Request is the options used to create the rpc HTTP request.
	// +optional
	Request *RequestOpts `json:"request"`
	// Signature is the options used for adding an HMAC signature to an HTTP request.
	// +optional
	Signature *SignatureOpts `json:"signature"`
	// HMAC is the options used to create a HMAC signature.
	// +optional
	HMAC *HMACOpts `json:"hmac"`
	// Experimental options.
	// +optional
	Experimental *ExperimentalOpts `json:"experimental"`
}

// RequestOpts are the options used when creating an HTTP request.
type RequestOpts struct {
	// HTTPContentType is the content type to use for the rpc request notification.
	// +optional
	HTTPContentType string `json:"httpContentType"`
	// HTTPMethod is the HTTP method to use for the rpc request notification.
	// +optional
	HTTPMethod string `json:"httpMethod"`
	// StaticHeaders are predefined headers that will be added to every request.
	// +optional
	StaticHeaders http.Header `json:"staticHeaders"`
	// TimestampFormat is the time format for the timestamp header.
	// +optional
	TimestampFormat string `json:"timestampFormat"`
	// TimestampHeader is the header name that should contain the timestamp. Example: X-BMCLIB-Timestamp
	// +optional
	TimestampHeader string `json:"timestampHeader"`
}

// SignatureOpts are the options used for adding an HMAC signature to an HTTP request.
type SignatureOpts struct {
	// HeaderName is the header name that should contain the signature(s). Example: X-BMCLIB-Signature
	// +optional
	HeaderName string `json:"headerName"`
	// AppendAlgoToHeaderDisabled decides whether to append the algorithm to the signature header or not.
	// Example: X-BMCLIB-Signature becomes X-BMCLIB-Signature-256
	// When set to true, a header will be added for each algorithm. Example: X-BMCLIB-Signature-256 and X-BMCLIB-Signature-512
	// +optional
	AppendAlgoToHeaderDisabled bool `json:"appendAlgoToHeaderDisabled"`
	// IncludedPayloadHeaders are headers whose values will be included in the signature payload. Example: X-BMCLIB-My-Custom-Header
	// All headers will be deduplicated.
	// +optional
	IncludedPayloadHeaders []string `json:"includedPayloadHeaders"`
}

// HMACOpts are the options used to create an HMAC signature.
type HMACOpts struct {
	// PrefixSigDisabled determines whether the algorithm will be prefixed to the signature. Example: sha256=abc123
	// +optional
	PrefixSigDisabled bool `json:"prefixSigDisabled"`
	// Secrets are a map of algorithms to secrets used for signing.
	// +optional
	Secrets HMACSecrets `json:"secrets"`
}

// ExperimentalOpts are options we're still learning about and should be used carefully.
type ExperimentalOpts struct {
	// CustomRequestPayload must be in json.
	// +optional
	CustomRequestPayload string `json:"customRequestPayload"`
	// DotPath is the path to the json object where the bmclib RequestPayload{} struct will be embedded. For example: object.data.body
	// +optional
	DotPath string `json:"dotPath"`
}
