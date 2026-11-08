package bot

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/kirugan/telegram-bot-api/v5"
)

func showMainMenu(bot *tgbotapi.BotAPI, chatID, userID int64) {
	// текст приветрвенного сообщения
	result := fmt.Sprintf(
		"Добро пожаловать в BH_Music!\n\n" +
			"Вы можете ознакомьтесь с документацией и пользовательским соглашением в README" +
			"Чтобы продолжить, примите соглашение и подпишитесь на канал проекта.",
	)

	ctxGUS, cancleGUS := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancleGUS()
	_, _, err := us_db.GetUserStatus(ctxGUS, userID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("error whem getting user status: %v", err)
		sendErrorMessage(bot, chatID, "Internal service error_1: повторите попытку позже", "")
		return
	} else if err == sql.ErrNoRows {
		ctxIU, cancleIU := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancleIU()
		err := us_db.InsertUser(ctxIU, userID, chatID, "nuser")
		if err != nil {
			log.Printf("error when insert user: %v", err)
			sendErrorMessage(bot, chatID, "Internal service error_2: повторите попытку позже", "")
			return
		}
		sendSuccessMessage(
			bot,
			chatID,
			result,
			"reg",
			0,
		)
	} else {
		sendSuccessMessage(
			bot,
			chatID,
			result,
			"reg",
			0,
		)
	}
}
