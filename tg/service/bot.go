package bot

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	config "github.com/nk-BH-D/BH_Music/tg/internal/config"
	mus "github.com/nk-BH-D/BH_Music/tg/internal/music_db"
	us "github.com/nk-BH-D/BH_Music/tg/internal/users_db"
)

var (
	userStates = make(map[int64]map[string]string)
	uSM        sync.RWMutex
	conf       *config.Config
	us_db      *us.PostgresUs
	mus_db     *mus.PostgresMus
)

func Init(cfg *config.Config, pg_mus *mus.PostgresMus, pg_us *us.PostgresUs) {
	conf = cfg
	us_db = pg_us
	mus_db = pg_mus
}

// func fot syncing userState map
func setUserState(chatID int64, state map[string]string) {
	uSM.Lock()
	defer uSM.Unlock()
	userStates[chatID] = state
}
func getUserState(chatID int64) (map[string]string, bool) {
	uSM.RLock()
	defer uSM.RUnlock()
	state, ok := userStates[chatID]
	return state, ok
}
func deleteUserState(chatID int64) {
	uSM.Lock()
	defer uSM.Unlock()
	delete(userStates, chatID)
}

func sendErrorMessage(bot *tgbotapi.BotAPI, chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	_, sendErr := bot.Send(msg)
	if sendErr != nil {
		log.Printf("error sending message: %s, %v", message, sendErr)
		return
	}
}
func sendSuccessMessage(bot *tgbotapi.BotAPI, chatID int64, message string, buttom string) {
	message_list := []string{}
	if len(message) > 4096 {
		runes := []rune(message)
		for i := 0; i < len(runes); i += 4096 {
			end := i + 4096

			if end > len(runes) {
				end = len(runes) // хвост
			}

			message_list = append(message_list, string(runes[i:end]))
		}
		assistedSendSuccessMessage(bot, chatID, message_list, buttom)
	} else {
		message_list = append(message_list, message)
		assistedSendSuccessMessage(bot, chatID, message_list, buttom)
	}
}
func assistedSendSuccessMessage(bot *tgbotapi.BotAPI, chatID int64, message_list []string, buttom string) {
	for _, message := range message_list {
		msg := tgbotapi.NewMessage(chatID, message)
		switch buttom {
		case "rating":
			msg.ReplyMarkup = createRatingKeeydoard()
		case "menu":
			msg.ReplyMarkup = createMenuKeyboard()
		}
		_, sendErr := bot.Send(msg)
		if sendErr != nil {
			log.Printf("error sending message: %s, %v", message, sendErr)
			sendErrorMessage(bot, chatID, fmt.Sprintf("произошла ошибка при отправке сообщения: %v", sendErr))
			log.Printf("ошибка при отправкее сообщения: %v", sendErr)
			return
		}
	}

}

func HandleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	userID := message.From.ID
	login := message.From.UserName
	full_name := strings.TrimSpace(fmt.Sprintf("%s %s", message.From.FirstName, message.From.LastName))

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Intercepted panic: %v", r)
			sendErrorMessage(bot, chatID, "Что-то пошло не так, попробуйте снова")
			return
		}
	}()

	if message.Text != "" && message.Text == "/start" {
		deleteUserState(chatID)
		go showMainMenu(bot, chatID, userID, login, full_name)
		log.Printf("Активныч горутин: %d\n", runtime.NumGoroutine())
		return
	}
}
