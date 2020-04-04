package command

type Category string

const (
	General         Category = "ℹ️ General"
	Tickets         Category = "📩 Tickets"
	Settings        Category = "🔧 Settings"
	CannedResponses Category = "✍️ Canned Responses"
	Statistics      Category = "📈 Statistics"
)

var categories = []Category{
	General,
	Tickets,
	Settings,
	CannedResponses,
	Statistics,
}
