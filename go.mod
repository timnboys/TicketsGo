module github.com/TicketsBot/TicketsGo

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/certifi/gocertifi v0.0.0-20200211180108-c7c1fbc02894 // indirect
	github.com/elliotchance/orderedmap v1.2.1
	github.com/getsentry/raven-go v0.2.0
	github.com/go-errors/errors v1.0.1
	github.com/go-redis/redis v6.15.7+incompatible
	github.com/go-sql-driver/mysql v1.5.0
	github.com/jinzhu/gorm v1.9.12
	github.com/jwangsadinata/go-multimap v0.0.0-20190620162914-c29f3d7f33b6
	github.com/onsi/ginkgo v1.12.0 // indirect
	github.com/onsi/gomega v1.9.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rxdn/gdl v0.0.0-20200405204544-572367093c4f
	github.com/satori/go.uuid v1.2.0
	gopkg.in/alexcesaro/statsd.v2 v2.0.0
)

replace github.com/rxdn/gdl v0.0.0-20200404222358-486a4f578b16 => ./../../rxdn/gdl
