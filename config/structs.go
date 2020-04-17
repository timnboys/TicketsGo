package config

type (
	Config struct {
		AboutMessage string
		Bot          Bot
		Database     Database
		Metrics      Metrics
		Redis        Redis
		Cache        Cache
		Sentry       Sentry
	}

	Bot struct {
		Token                 string
		Prefix                string
		Admins                []uint64
		Helpers               []uint64
		PremiumLookupProxyUrl string `toml:"premium-lookup-proxy-url"`
		PremiumLookupProxyKey string `toml:"premium-lookup-proxy-key"`
		Sharding              Sharding
		Game                  string
		ObjectStore           string
	}

	Sharding struct {
		Total  int
		Lowest int
		Max    int
	}

	Database struct {
		Host     string
		Port     int
		Username string
		Password string
		Database string
		Pool     Pool
		Lifetime int
	}

	Pool struct {
		MaxConnections int
		MaxIdle        int
	}

	Metrics struct {
		Statsd Statsd
	}

	Statsd struct {
		Enabled bool
		Prefix  string
		Host    string
		Port    int
	}

	Redis struct {
		Enabled bool
		Uri     string
		Threads int
	}

	Cache struct {
		Uri string
	}

	Sentry struct {
		DSN string
	}
)
