package bot

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func showMainMenu(bot *tgbotapi.BotAPI, chatID, userID int64, login, name string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("sMM: intercepted panic: %v", r)
			sendErrorMessage(bot, chatID, "Что-то пошло не так, попробуйте снова", "menu")
			return
		}
	}()

	ctxGUS, cancleGUS := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancleGUS()
	_, _, _, err := us_db.GetUserStatus(ctxGUS, userID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("error whem getting user status: %v", err)
		sendErrorMessage(bot, chatID, "Internal service error_1: повторите попытку позже", "")
		return
	} else if err == sql.ErrNoRows {
		ctxIU, cancleIU := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancleIU()
		err := us_db.InsertUser(ctxIU, userID, chatID, "nuser", login, name)
		if err != nil {
			log.Printf("error when insert user: %v", err)
			sendErrorMessage(bot, chatID, "Internal service error_2: повторите попытку позже", "")
			return
		}
		sendSuccessMessage(
			bot,
			chatID,
			fmt.Sprintf(
				"Привет %s,\nВы можете подписаться на канал автора %s и ознакомиться с кодом и инструкцией тут.\nGitHub: %s\nЧто-бы продолжить нажмите 'Меню'.",
				name,
				conf.TG_CHANNEL_URL,
				conf.REPO_URL,
			),
			"menu",
		)
	} else {
		sendSuccessMessage(
			bot,
			chatID,
			fmt.Sprintf(
				"Привет %s,\nВы можете подписаться на канал автора %s и ознакомиться с кодом и инструкцией тут. %s",
				name,
				conf.TG_CHANNEL_URL,
				conf.REPO_URL,
			),
			"menu",
		)
	}
}
