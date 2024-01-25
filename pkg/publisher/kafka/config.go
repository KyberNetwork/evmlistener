package kafka

type Config struct {
	Addresses []string

	UseAuthentication bool
	Username          string
	Password          string
}
