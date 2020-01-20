package listeners

import (
	"errors"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/bwmarrin/discordgo"
	"strconv"
)

func OnUserUpdate(s *discordgo.Session, e *discordgo.UserUpdate) {
	id, err := strconv.ParseInt(e.User.ID, 10, 64); if err != nil {
		sentry.ErrorWithContext(err, sentry.ErrorContext{
			User:  e.User.ID,
			Shard: s.ShardID,
		})
		return
	}

	go database.UpdateUser(id, e.Username, e.Discriminator, e.Avatar)
}

func OnUserJoin(s *discordgo.Session, e *discordgo.GuildMemberAdd) {
	id, err := strconv.ParseInt(e.User.ID, 10, 64); if err != nil {
		sentry.ErrorWithContext(err, sentry.ErrorContext{
			Guild: e.GuildID,
			User:  e.User.ID,
			Shard: s.ShardID,
		})
		return
	}

	go database.UpdateUser(id, e.User.Username, e.User.Discriminator, e.User.Avatar)
}

func OnGuildCreateUserData(_ *discordgo.Session, e *discordgo.GuildCreate) {
	data := make([]database.UserData, 0)

	for _, member := range e.Members {
		userId, err := strconv.ParseInt(member.User.ID, 10, 64); if err != nil {
			sentry.Error(errors.New(err.Error()))
			continue
		}

		data = append(data, database.UserData{
			UserId:        userId,
			Username:      member.User.Username,
			Discriminator: member.User.Discriminator,
			Avatar:        member.User.Avatar,
		})
	}

	go database.InsertUsers(data)
}
