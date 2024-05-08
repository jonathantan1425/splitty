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

const splitTag = "[Split] "

const splitInfo = splitTag + "Let's split a bill. Please enter the amount that you would like to split equally followed by mentioning the people you would like to split the bill with."

const INSERT_SQL_STMT = "INSERT INTO owing_history (chat_id, debtor_id, creditor_id, amount) values(?, ?, ?, ?)"

func SplitHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	slog.Info("/split command received from chat %d", update.Message.Chat.ID)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   splitInfo,
		ReplyMarkup: models.ForceReply{
			ForceReply: true,
		},
	})
}

func SplitReplyMatchHandler(update *models.Update) bool {
	slog.Debug("ReplyHandler received from chat %d", update.Message.Chat.ID)
	if update.Message.ReplyToMessage == nil {
		return false
	}

	if strings.Contains(update.Message.ReplyToMessage.Text, splitTag) {
		return true
	}

	return false
}

func SplitReplyHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	slog.Info("splitreply received from chat %d", update.Message.Chat.ID)

	if update.Message.ReplyToMessage == nil {
		return
	}

	if update.Message.Text == "" {
		slog.Error("No amount entered")
		return
	}

	text := strings.Fields(update.Message.Text)
	if len(text) < 2 {
		slog.Error("Not enough fields in message")
		return
	}

	// Check if the amount is a valid number
	amount, err := strconv.ParseFloat(text[0], 64)
	if err != nil {
		slog.Error("Error parsing amount: %v", err)
		return
	}

	// Split the amount among all participants
	members := len(update.Message.Entities)
	slog.Info("Number of members in chat: %d", members)

	// Add the sender as a participant
	membersWithoutBot := members + 1

	// Split the amount equally among all members
	amountPerPerson := amount / float64(membersWithoutBot)
	slog.Info("Amount per person: %s", amountPerPerson)

	// Log the amount per person
	db, dbError = GetDB()
	if dbError != nil {
		slog.Error("Error getting database: %v", dbError)
		return
	}
	if db == nil {
		slog.Error("Database is nil")
		return
	}

	insertStmt, err := db.Prepare(INSERT_SQL_STMT)

	if err != nil {
		slog.Error("Error preparing insert statement: %v", err)
		return
	}

	userRow := db.QueryRow("SELECT id FROM chat_user_identities WHERE chat_id = ? AND (username = ? OR user_id = ?)", update.Message.Chat.ID, update.Message.From.Username, update.Message.From.ID)
	var creditorId int
	err = userRow.Scan(&creditorId)
	if err != nil {
		if err != sql.ErrNoRows {
			slog.Error("Error scanning user row: %v", err)
			return
		} else {
			slog.Info("No existing user row found, inserting new row")
			_, err = db.Exec("INSERT INTO chat_user_identities (chat_id, username, user_id) values(?, ?, ?)", update.Message.Chat.ID, update.Message.From.Username, update.Message.From.ID)
			if err != nil {
				slog.Error("Error inserting user row: %v", err)
				return
			}
			newRow := db.QueryRow("SELECT id FROM chat_user_identities WHERE chat_id = ? AND username = ? AND user_id = ?", update.Message.Chat.ID, update.Message.From.Username, update.Message.From.ID)
			err = newRow.Scan(&creditorId)
		}
	}

	for _, entity := range update.Message.Entities {
		fmt.Printf("entity: %v\n", entity)

		var debtorId int

		if entity.Type == "mention" {
			debtorUsername := update.Message.Text[entity.Offset+1:entity.Offset+entity.Length]

			userRow := db.QueryRow("SELECT id FROM chat_user_identities WHERE chat_id = ? AND username = ?", update.Message.Chat.ID, debtorUsername)
			err = userRow.Scan(&debtorId)
			if err != nil {
				if err != sql.ErrNoRows {
					slog.Error("Error scanning user row: %v", err)
					return
				} else {
					slog.Info("No existing user row found, inserting new row")
					_, err = db.Exec("INSERT INTO chat_user_identities (chat_id, username) values(?, ?)", update.Message.Chat.ID, debtorUsername)
					if err != nil {
						slog.Error("Error inserting user row: %v", err)
						return
					}
					newRow := db.QueryRow("SELECT id FROM chat_user_identities WHERE chat_id = ? AND username = ?", update.Message.Chat.ID, debtorUsername)
					err = newRow.Scan(&debtorId)
					if err != nil {
						slog.Error("Error scanning new user row: %v", err)
						return
					}
				}
			}
		} else if entity.Type == "text_mention" {
			debtorUsername := entity.User.Username
			debtorUserId := int(entity.User.ID)

			userRow := db.QueryRow("SELECT * FROM chat_user_identities WHERE chat_id = ? AND username = ? AND user_id = ?", update.Message.Chat.ID, debtorUsername, debtorUserId)
			var chatId, userId int
			var username string
			err = userRow.Scan(&debtorId, &chatId, &username, &userId)
			if err != nil {
				if err != sql.ErrNoRows {
					slog.Error("Error scanning user row: %v", err)
					return
				} else {
					slog.Info("No existing user row found, inserting new row")
					_, err = db.Exec("INSERT INTO chat_user_identities (chat_id, username, user_id) values(?, ?, ?)", update.Message.Chat.ID, debtorUsername, debtorUserId)
					if err != nil {
						slog.Error("Error inserting user row: %v", err)
						return
					}
					newRow := db.QueryRow("SELECT id FROM chat_user_identities WHERE chat_id = ? AND username = ? AND user_id = ?", update.Message.Chat.ID, debtorUsername, debtorUserId)
					err = newRow.Scan(&debtorId)
				}
			}
		}

		if debtorId != 0 {
			row := db.QueryRow("SELECT amount FROM owing_history WHERE chat_id = ? AND debtor_id = ? AND creditor_id = ?", update.Message.Chat.ID, debtorId, creditorId)
			var currentAmount float64
			err = row.Scan(&currentAmount)
			if err != nil {
				if err != sql.ErrNoRows {
					slog.Error("Error scanning row: %v", err)
					return
				} else {
					slog.Info("No existing row found, inserting new row")
					_, err = insertStmt.Exec(update.Message.Chat.ID, debtorId, creditorId, amountPerPerson)
				}
			} else {
				slog.Info("Existing row found, updating row")
				_, err = db.Exec("UPDATE owing_history SET amount = amount + ? WHERE chat_id = ? AND debtor_id = ? AND creditor_id = ?", amountPerPerson, update.Message.Chat.ID, debtorId, creditorId)
			}
		}
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:  	"Added amount per person: " + strconv.FormatFloat(amountPerPerson, 'f', 2, 64),
	})
	if err != nil {
		slog.Error("Error sending message: %v", err)
	}
}
