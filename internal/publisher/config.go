package publisher

type Config struct {
	Topic       string `json:"topic"`
	OrderingKey string `json:"orderingKey"`
}
