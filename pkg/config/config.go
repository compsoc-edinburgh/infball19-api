package config

type Config struct {
	LogLevel string `default:"debug"`

	// Stripe
	Stripe StripeConfig

	// Bind Address
	BindAddress string `default:"0.0.0.0:8080"`

	// Mailgun
	Mailgun MailgunConfig

	// Pass
	StatsPass string `required:"true"`

	// Staff Code
	StaffCode string `required:"true"`

	// Public website url
	PublicURL string `required:"true"`

	// Redis connection parameters
	Redis RedisConfig

	// Endpoint for qr codes
	QRLocation string `required:"true"`
}

// StripeConfig contains all configuration data for a CoSign connection
type StripeConfig struct {
	PublishableKey  string `required:"true"`
	SecretKey       string `required:"true"`
	Product         string `required:"false"` // stripe product
	SKU             string `required:"true"`  // stripe SKU
	NonAlcoholicSKU string `required:"true"`  // SKU for non alcoholic ticket
}

type RedisConfig struct {
	Address  string `default: "localhost:6739"`
	Password string `default: ""`
	DB       int    `default:"0"`
}

// Token is an "API" user
type Token struct {
	Name string `required:"true"`
	Key  string `required:"true"`
}

type MailgunConfig struct {
	Domain       string `required:"true"`
	APIKey       string `required:"true"`
	PublicAPIKey string `required:"true"`
}
