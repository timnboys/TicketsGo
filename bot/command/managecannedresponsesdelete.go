package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/objects/channel/embed"
)

type ManageTagsDelete struct {
}

func (ManageTagsDelete) Name() string {
	return "delete"
}

func (ManageTagsDelete) Description() string {
	return "Deletes a tag"
}

func (ManageTagsDelete) Aliases() []string {
	return []string{"del", "rm", "remove"}
}

func (ManageTagsDelete) PermissionLevel() utils.PermissionLevel {
	return utils.Support
}

func (ManageTagsDelete) Execute(ctx utils.CommandContext) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!managetags delete [TagID]`",
		Inline: false,
	}

	if len(ctx.Args) == 0 {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", "You must specify a tag ID to delete", usageEmbed)
		return
	}

	id := ctx.Args[0]

	var found bool
	{
		tag, err := database.Client.Tag.Get(ctx.GuildId, id)
		if err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			ctx.ReactWithCross()
			return
		}

		found = tag != ""
	}

	if !found {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", fmt.Sprintf("A tag with the ID `%s` could not be found", id))
		return
	}

	if err := database.Client.Tag.Delete(ctx.GuildId, id); err == nil {
		ctx.ReactWithCheck()
	} else {
		ctx.ReactWithCross()
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}
}

func (ManageTagsDelete) Parent() interface{} {
	return ManageTags{}
}

func (ManageTagsDelete) Children() []Command {
	return make([]Command, 0)
}

func (ManageTagsDelete) PremiumOnly() bool {
	return false
}

func (ManageTagsDelete) Category() Category {
	return Tags
}

func (ManageTagsDelete) AdminOnly() bool {
	return false
}

func (ManageTagsDelete) HelperOnly() bool {
	return false
}
