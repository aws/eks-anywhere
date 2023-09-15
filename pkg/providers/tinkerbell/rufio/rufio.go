package rufio

// Config is for Rufio specific configuration.
type Config struct {
	// WebhookSecret is the secret that will be used to sign webhook notifications in Rufio.
	// Multiple secrets can be used by separating them with a comma.
	WebhookSecret string
	// ConsumerURL is the URL that will be used when sending webhook notifications in Rufio.
	ConsumerURL string
}
