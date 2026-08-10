package kafka

type Config struct {
	Addresses []string

	UseAuthentication bool
	// SASLMechanism is the SASL mechanism to use when UseAuthentication is true.
	// Supported values: "PLAIN" (default), "SCRAM-SHA-512".
	SASLMechanism string
	Username      string
	Password      string
}
