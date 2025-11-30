package config

type Config struct {
	Server        ServerConfig
	HttpClient    HttpClientConfig
	WebhookConfig WebhookConfig
	Scheduler     SchedulerConfig
	Database      DatabaseConfig
	Redis         RedisConfig
}

type ServerConfig struct {
	AppVersion   string      `split_words:"true"`
	HttpPort     string      `required:"true" split_words:"true" default:"8080"`
	Environment  Environment `required:"true" split_words:"true" default:"local"`
	LogLevel     string      `split_words:"true" default:"INFO"`
	ReadTimeout  int         `split_words:"true" default:"5"`
	WriteTimeout int         `split_words:"true" default:"10"`
}

type HttpClientConfig struct {
	Timeout       int `split_words:"true" default:"5"`
	MaxConnection int `split_words:"true" default:"5"`
}

type WebhookConfig struct {
	Url     string `split_words:"true" default:"http://localhost:9000/webhook"`
	AuthKey string `split_words:"true"`
}

type SchedulerConfig struct {
	IntervalInSeconds int `split_words:"true" default:"120"`
	BatchSize         int `split_words:"true" default:"2"`
}

type DatabaseConfig struct {
	Name string `split_words:"true" default:"message"`
}

type RedisConfig struct {
	Host         string `split_words:"true" default:"localhost"`
	Port         string `split_words:"true" default:"6379"`
	Password     string `split_words:"true" default:""`
	DB           int    `split_words:"true" default:"0"`
	TTLInSeconds int    `split_words:"true" default:"3600"`
}
