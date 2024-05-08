package main

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const helpInfo = `
List of commands:
- /split - Add a bill to split among selected participants in the chat
- /balance - See outstanding bills in this chat
- /settle - Settle outstanding bills in this chat
`

func HelpHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	slog.Debug("Help command received from chat %d", update.Message.Chat.ID)
	slog.Info("Help command received from chat %d", update.Message.Chat.ID)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      helpInfo,
	})
}
