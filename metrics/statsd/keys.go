package statsd

type Key string

const (
	MESSAGES Key = "messages"
	TICKETS Key = "tickets"
)

func (k Key) String() string {
	return string(k)
}
