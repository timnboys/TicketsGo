package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"regexp"
	"strconv"
)

var timeRegex = regexp.MustCompile(`(?:(?P<months>\d+)M)*(?:(?P<weeks>\d+)w)*(?:(?P<days>\d+)d)*(?:(?P<hours>\d+)h)*(?:(?P<minutes>\d+)m)*`)

type PurgeCommand struct {
}

func (PurgeCommand) Name() string {
	return "purge"
}

func (PurgeCommand) Description() string {
	return "Automatically close tickets which have not received a response in a certain amount of time"
}

func (PurgeCommand) Aliases() []string {
	return []string{}
}

func (PurgeCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Support
}

func (PurgeCommand) Execute(ctx utils.CommandContext) {
	/*if len(ctx.Args) == 0 {
		ctx.SendEmbed(utils.Red, "Error", `You must specify a length of time for which tickets must not have had a reply to purge.
M = Month, W = Week, d = Day, h = Hour, m = Minute
For example, to mute someone for 1 month, 3 weeks and 2 hours, use the time period of `+"`1M3W2h`")
		return
	}

	reason := "Ticket is inactive"
	if len(ctx.Args) > 1 {
		reason = strings.Join(ctx.Args[1:len(ctx.Args)], " ")
	}

	// Parse user input
	period := ctx.Args[0]

	var months, weeks, days, hours, minutes int

	res := timeRegex.FindAllStringSubmatch(period, -1)
	for index, value := range res[0] {
		switch index {
		case 1:
			months = positiveIntOrZero(value)
		case 2:
			weeks = positiveIntOrZero(value)
		case 3:
			days = positiveIntOrZero(value)
		case 4:
			hours = positiveIntOrZero(value)
		case 5:
			minutes = positiveIntOrZero(value)
		}
	}

	before := time.Now()
	before.Add(time.Duration(-minutes) * time.Minute)
	before.Add(time.Duration(-hours) * time.Hour)
	before.Add(time.Duration(-days) * time.Hour * 24)
	before.Add(time.Duration(-weeks) * time.Hour * 24 * 7)
	before.Add(time.Duration(-months) * time.Hour * 24 * 7 * 4)

	ticketsChan := make(chan []database.Ticket)
	go database.GetOpenTicketStructs(ctx.GuildId, ticketsChan)
	tickets := <-ticketsChan

	ctx.SendEmbed(utils.Green, "Purge", "Processing... This may take a while.")

	for _, ticket := range tickets {
		channelId := strconv.Itoa(int(*ticket.Channel))

		msgs, err := ctx.Shard.ChannelMessages(channelId, 1, "", "", "")
		if err != nil || len(msgs) == 0 { // Shouldn't ever happen?
			continue
		}

		lastMsg := msgs[0]
		time, err := lastMsg.Timestamp.Parse()
		if err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			continue
		}

		if before.After(time) { // We should purge
			fakeContext := utils.CommandContext{
				Shard:     ctx.Shard,
				User:      ctx.Author,
				UserID:    ctx.AuthorID,
				Guild:     ctx.Guild,
				GuildId:   ctx.GuildId,
				Channel:   channelId,
				ChannelId: *ticket.Channel,
				Message:   ctx.Message,
				MessageId: ctx.MessageId,
				Root:      "close",
				Args:      strings.Split(reason, " "),
				IsPremium:   ctx.IsPremium,
				ShouldReact: ctx.ShouldReact,
				Member:      ctx.Member,
				IsFromPanel: false,
			}

			go CloseCommand{}.Execute(fakeContext)
		}
	}

	ctx.SendEmbed(utils.Green, "Purge", "Purge has been completed")*/
}

// If err != nil, i = 0
func positiveIntOrZero(s string) int {
	i, _ := strconv.ParseInt(s, 10, 32)

	if i < 0 {
		i = 0
	}

	return int(i)
}

func (PurgeCommand) Parent() interface{} {
	return nil
}

func (PurgeCommand) Children() []Command {
	return make([]Command, 0)
}

func (PurgeCommand) PremiumOnly() bool {
	return false
}

func (PurgeCommand) Category() Category {
	return Settings
}

func (PurgeCommand) AdminOnly() bool {
	return false
}

func (PurgeCommand) HelperOnly() bool {
	return false
}
