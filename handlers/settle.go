package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const settleTag = "[Settle] "

const settleInfo = settleTag + "Let's settle a bill. Please enter the amount that you would like to settle followed by mentioning the person you would like to settle the bill with."

func SettleHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	slog.Info("/settle command received from chat %d", update.Message.Chat.ID)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   settleInfo,
		ReplyMarkup: models.ForceReply{
			ForceReply: true,
		},
	})
}

func SettleReplyMatchHandler(update *models.Update) bool {
	slog.Debug("ReplyHandler received from chat %d", update.Message.Chat.ID)
	if update.Message.ReplyToMessage == nil {
		return false
	}

	if strings.Contains(update.Message.ReplyToMessage.Text, settleTag) {
		return true
	}

	return false
}

func SettleReplyHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	slog.Info("settlereply received from chat %d", update.Message.Chat.ID)

	if update.Message.ReplyToMessage == nil {
		return
	}

	if update.Message.Text == "" {
		slog.Error("No amount entered")
		return
	}

	text := strings.Fields(update.Message.Text)
	if len(text) != 2 {
		slog.Error("Wrong number of fields in message")
		return
	}

	amount, err := strconv.ParseFloat(text[0], 64)
	if err != nil {
		slog.Error("Error parsing amount: %v", err)
		return
	}

	// Get the chat ID
	chatID := update.Message.Chat.ID

	senderId := update.Message.From.ID
	senderUsername := update.Message.From.Username

	userRow := db.QueryRow("SELECT id FROM chat_user_identities WHERE chat_id = ? AND (username = ? OR user_id = ?)", chatID, senderUsername, senderId)
	var debtorId int
	err = userRow.Scan(&debtorId)
	if err != nil {
		slog.Error("Error getting user ID: %v", err)
		return
	}

	if len(update.Message.Entities) == 0 {
		slog.Error("No participants mentioned")
		return
	} else if len(update.Message.Entities) > 1 {
		slog.Error("Too many participants mentioned")
		return
	}

	// Get the user ID of the participant
	entity := update.Message.Entities[0]
	var creditorUsername string
	var creditorUserId int

	var creditorId int
	if entity.Type == "mention" {
		creditorUsername = update.Message.Text[entity.Offset+1 : entity.Offset+entity.Length]

		userRow = db.QueryRow("SELECT id FROM chat_user_identities WHERE chat_id = ? AND username = ?", chatID, creditorUsername)
		err = userRow.Scan(&creditorId)
		if err != nil {
			slog.Error("Error getting creditor ID: %v", err)
			return
		}
	} else if entity.Type == "text_mention" {
		creditorUsername = entity.User.Username
		creditorUserId = int(entity.User.ID)

		userRow = db.QueryRow("SELECT id FROM chat_user_identities WHERE chat_id = ? AND (username = ? OR user_id = ?)", chatID, creditorUsername, creditorUserId)
		err = userRow.Scan(&creditorId)
		if err != nil {
			slog.Error("Error getting creditor ID: %v", err)
			return
		}
	}

	if creditorId == 0 {
		slog.Error("No creditor ID found")
		return
	} else if creditorId == debtorId {
		slog.Error("Creditor and debtor are the same")
		return
	}

	// Check outstanding amount owed
	var totalOwed float64
	row := db.QueryRow("SELECT amount FROM owing_history WHERE chat_id = ? AND debtor_id = ? AND creditor_id = ?", chatID, debtorId, creditorId)
	err = row.Scan(&totalOwed)
	if err != nil {
		if err != sql.ErrNoRows {
			slog.Error("Error scanning row: %v", err)
			return
		} else {
			slog.Error("No outstanding amount found")
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "No outstanding owed amount found",
			})
			return
		}
	}

	if amount > totalOwed {
		slog.Error("Amount exceeds total owed")
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Amount exceeds total owed, please enter an amount less than or equal to $%.2f", totalOwed),
		})
		return
	}

	// Update the amount owed
	_, err = db.Exec("UPDATE owing_history SET amount = amount - ? WHERE chat_id = ? AND debtor_id = ? AND creditor_id = ?", amount, chatID, debtorId, creditorId)
	if err != nil {
		slog.Error("Error updating amount: %v", err)
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Amount of %.2f settled between %s and %s", amount, BuildUserMention(senderUsername, senderId), BuildUserMention(creditorUsername, int64(creditorUserId))),
	})
}
