package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/rxdn/gdl/objects/channel/embed"
	"strings"
)

type ManageTagsAdd struct {
}

func (ManageTagsAdd) Name() string {
	return "add"
}

func (ManageTagsAdd) Description() string {
	return "Adds a new tag"
}

func (ManageTagsAdd) Aliases() []string {
	return []string{"new", "create"}
}

func (ManageTagsAdd) PermissionLevel() utils.PermissionLevel {
	return utils.Support
}

func (ManageTagsAdd) Execute(ctx utils.CommandContext) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!managetags add [TagID] [Tag contents]`",
		Inline: false,
	}

	if len(ctx.Args) < 2 {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", "You must specify a tag ID and contents", usageEmbed)
		return
	}

	id := ctx.Args[0]
	content := ctx.Args[1:] // content cannot be bigger than the Discord limit, obviously

	// Length check
	if len(id) > 16 {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", "Tag IDs cannot be longer than 16 characters", usageEmbed)
		return
	}

	// Verify a tag with the ID doesn't already exist
	tagExists := make(chan bool)
	go database.CannedResponseExists(ctx.GuildId, id, tagExists)
	if <-tagExists {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", "A tag with the ID `$id` already exists. You can delete the response using `t!managetags delete [ID]`", usageEmbed)
		return
	}

	go database.AddCannedResponse(ctx.GuildId, id, strings.Join(content, " "))
	ctx.ReactWithCheck()
}

func (ManageTagsAdd) Parent() interface{} {
	return ManageTags{}
}

func (ManageTagsAdd) Children() []Command {
	return make([]Command, 0)
}

func (ManageTagsAdd) PremiumOnly() bool {
	return false
}

func (ManageTagsAdd) Category() Category {
	return Tags
}

func (ManageTagsAdd) AdminOnly() bool {
	return false
}

func (ManageTagsAdd) HelperOnly() bool {
	return false
}
