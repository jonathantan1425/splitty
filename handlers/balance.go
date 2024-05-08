package main

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"database/sql"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const QUERY_SQL_STMT = `WITH OUTSTANDING_BALANCES AS (
	SELECT
	o.debtor_id,
	o.creditor_id,
	chat_id,
	SUM(amount) -
	(SELECT COALESCE(SUM(amount), 0) FROM owing_history
		WHERE chat_id = o.chat_id AND creditor_id = o.debtor_id AND debtor_id = o.creditor_id) AS total_owed
	FROM owing_history o
	WHERE chat_id = ?
	GROUP BY debtor_id, creditor_id
	HAVING total_owed > 0)
	
	SELECT
	cu1.user_id AS debtor_id,
	cu1.username AS debtor_username,
	cu2.user_id AS creditor_id,
	cu2.username AS creditor_username,
	total_owed
	FROM OUTSTANDING_BALANCES o
	JOIN chat_user_identities cu1 ON o.debtor_id = cu1.id AND o.chat_id = cu1.chat_id
	JOIN chat_user_identities cu2 ON o.creditor_id = cu2.id AND o.chat_id = cu2.chat_id
	ORDER BY debtor_id, creditor_id
	;
`

func BalanceHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	slog.Info("/balance command received from chat %d", update.Message.Chat.ID)

	db, dbError = GetDB()
	if dbError != nil {
		slog.Error("Error getting database: %v", dbError)
		return
	}
	if db == nil {
		slog.Error("Database is nil")
		return
	}

	stmt, err := db.Prepare(QUERY_SQL_STMT)
	if err != nil {
		slog.Error("Error preparing statement: %v", err)
		return
	}

	rows, err := stmt.Query(update.Message.Chat.ID)
	if err != nil {
		slog.Error("Error querying database: %v", err)
		return
	}
	defer rows.Close()

	var balanceMessage string

	for rows.Next() {
		var debtorId, creditorId sql.NullInt64
		var debtorUsername, creditorUsername string
		var totalOwed float64
		err = rows.Scan(&debtorId, &debtorUsername, &creditorId, &creditorUsername, &totalOwed)
		if err != nil {
			slog.Error("Error scanning row: %v", err)
			return
		}

		balanceMessage += fmt.Sprintf(
			"%s owes %s %s\n",
			BuildUserMention(debtorUsername, debtorId.Int64),
			BuildUserMention(creditorUsername, creditorId.Int64),
			bot.EscapeMarkdown(strconv.FormatFloat(totalOwed, 'f', 2, 64)),
		)
	}

	if balanceMessage == "" {
		balanceMessage = "No outstanding balances"
	}

	slog.Info("Balance message: %s", balanceMessage)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      balanceMessage,
		ParseMode: models.ParseModeMarkdown,
	})
}

func BuildUserMention(username string, userId int64) string {
	if username != "" {
		return bot.EscapeMarkdown("@" + username)
	} else {
		return "[User](tg://user?id=" + fmt.Sprintf("%d", userId) + ")"
	}
}
