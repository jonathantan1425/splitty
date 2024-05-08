package main

import (
	"context"
	"os"
	"os/signal"
	"log/slog"
	"sync"
	"database/sql"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// Send any text message to the bot after the bot has been started

var (
	db   *sql.DB
	once  sync.Once
	dbError error
)

func main() {
	db, dbError = GetDB()
	if dbError != nil {
		slog.Error("Error opening database:", dbError)
	}
	defer db.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(DefaultHandler),
	}

	Env_init()
	token := os.Getenv("TELEGRAM_API_KEY")

	b, err := bot.New(token, opts...)
	if err != nil {
		panic(err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeExact, HelpHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/split", bot.MatchTypeExact, SplitHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/balance", bot.MatchTypeExact, BalanceHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/settle", bot.MatchTypeExact, SettleHandler)
	b.RegisterHandlerMatchFunc(SplitReplyMatchHandler, SplitReplyHandler)
	b.RegisterHandlerMatchFunc(SettleReplyMatchHandler, SettleReplyHandler)

	b.Start(ctx)
}

func DefaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
}
