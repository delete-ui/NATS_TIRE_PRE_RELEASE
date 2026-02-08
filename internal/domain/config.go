package domain

type MainSettings struct {
	AppName string `envconfig:"APP_NAME" default:"nats_tire_service"`
	Version string `envconfig:"VERSION"`
	Env     string `envconfig:"ENV" default:"production"`
}

type NATSTireServer struct {
	NATSPort     int    `envconfig:"NATS_PORT" default:"4222"`
	NATSHTTPPort int    `envconfig:"NATS_HTTP_PORT" default:"8222"`
	DataDir      string `envconfig:"DATA_DIR" default:"./data"`
}

type JetStreamSettings struct {
	JetStreamEnabled bool   `envconfig:"JETSTREAM_ENABLED" default:"true"`
	MaxMemoryStore   string `envconfig:"MAX_MEMORY_STORE" default:"1GB"`
	MaxFileStore     string `envconfig:"MAX_FILE_STORE" default:"10GB"`
}

type NATSSecurity struct {
	NATSUsername string `envconfig:"NATS_USERNAME"`
	NATSPassword string `envconfig:"NATS_PASSWORD"`
}

type LoggerSettings struct {
	LogLevel  string `envconfig:"LOG_LEVEL" default:"INFO"`
	LogFile   string `envconfig:"LOG_FILE"`
	LogFormat string `envconfig:"LOG_FORMAT" default:"console"` //json/console
}

type HTTPServer struct {
	HTTPPort int `envconfig:"HTTP_PORT" default:"8080"`
}
