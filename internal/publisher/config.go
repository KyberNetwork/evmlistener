package publisher

type Config struct {
	Topic       string `json:"topic"`
	OrderingKey string `json:"orderingKey"`
}

type Type string

const (
	DataCenter  Type = "data-center"
	RedisStream Type = "redis-stream"
)

func (t Type) String() string {
	return string(t)
}
