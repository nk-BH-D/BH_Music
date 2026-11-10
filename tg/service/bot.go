package bot

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	tgbotapi "github.com/kirugan/telegram-bot-api/v5"
	config "github.com/nk-BH-D/BH_Music/tg/internal/config"
	db "github.com/nk-BH-D/BH_Music/tg/internal/music_users_db"
)

var (
	userStates = make(map[int64]map[string]string)
	broSlot    chan struct{}
	uSM        sync.RWMutex
	conf       *config.Config
	us_db      *db.PostgresUs
	mus_db     *db.PostgresMus
)

func Init(cfg *config.Config, pg_mus *db.PostgresMus, pg_us *db.PostgresUs) {
	conf = cfg
	broSlot = make(chan struct{}, cfg.MAX_BRO)
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

func sendErrorMessage(bot *tgbotapi.BotAPI, chatID int64, message, buttom string) {
	msg := tgbotapi.NewMessage(chatID, message)
	switch buttom {
	//case "rating":
	//	msg.ReplyMarkup = createRatingKeeydoard()
	case "menu":
		msg.ReplyMarkup = createMenuKeyboard()
	case "command":
		msg.ReplyMarkup = createCommandKeyboard()
	case "":
		msg.ReplyMarkup = nil
	default:
		log.Printf("unknown button type: %s", buttom)
		sendErrorMessage(bot, chatID, "Internal service error4: повторите попытку позже", "")
		return
	}
	_, sendErr := bot.Send(msg)
	if sendErr != nil {
		log.Printf("error sending message: %s, %v", message, sendErr)
		return
	}
}

func sendSuccessMessage(bot *tgbotapi.BotAPI, chatID int64, message, buttom string, msgID int) {
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
		assistedSendSuccessMessage(bot, chatID, message_list, buttom, msgID)
	} else {
		message_list = append(message_list, message)
		assistedSendSuccessMessage(bot, chatID, message_list, buttom, msgID)
	}
}
func assistedSendSuccessMessage(bot *tgbotapi.BotAPI, chatID int64, message_list []string, buttom string, msgID int) {
	for _, message := range message_list {
		msg := tgbotapi.NewMessage(chatID, message)
		if msgID != 0 {
			msg.ReplyParameters = &tgbotapi.ReplyParameters{MessageID: msgID}
		}
		switch buttom {
		//case "rating":
		//	msg.ReplyMarkup = createRatingKeeydoard()
		case "menu":
			msg.ReplyMarkup = createMenuKeyboard()
		case "command":
			msg.ReplyMarkup = createCommandKeyboard()
		case "reg":
			msg.ReplyMarkup = createRegistrKeybord()
		case "":
			msg.ReplyMarkup = nil
		default:
			log.Printf("unknown button type: %s", buttom)
			sendErrorMessage(bot, chatID, "Internal service error3: повторите попытку позже", "")
			return
		}
		_, sendErr := bot.Send(msg)
		if sendErr != nil {
			log.Printf("error sending message: %s, %v", message, sendErr)
			sendErrorMessage(bot, chatID, fmt.Sprintf("произошла ошибка при отправке сообщения: %v", sendErr), "")
			log.Printf("ошибка при отправкее сообщения: %v", sendErr)
			return
		}
	}

}

// функция валидации пользователлей
func userValid(bot *tgbotapi.BotAPI, userID, chatID int64) bool {
	ctx, cancle := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancle()
	accept, _, err := us_db.GetUserStatus(ctx, userID)
	if err != nil {
		log.Printf("error when valid user: %v", err)
		sendErrorMessage(bot, chatID, "Произошла ошибка при проверке прав доступа, повторите попытку позже", "menu")
		return false
	}
	if !accept {
		sendErrorMessage(bot, chatID, "Вам отказано в доступе.\nВозможно вы не подписаны на канал проекта или не приняли соглашение", "")
		return false
	}

	var (
		fl bool = false
	)

	member, err := bot.GetChatMember(
		tgbotapi.GetChatMemberConfig{
			ChatID: conf.TG_CHANNEL_ID,
			UserID: userID,
		},
	)
	if err != nil {
		log.Printf("error when chek sub: %v", err)
		sendErrorMessage(bot, chatID, "Internal service error14: повторите попытку позже", "")
		return false
	} else {
		switch member.Status {
		case "creator", "administrator", "member":
			fl = true
		case "restricted":
			if member.IsMember {
				fl = true
			} else {
				sendErrorMessage(bot, chatID, "Вам отказано в доступе.\nВозможно вы не подписаны на канал проекта или не приняли соглашение", "")
				fl = false
			}
		default:
			sendErrorMessage(bot, chatID, "Вам отказано в доступе.\nВозможно вы не подписаны на канал проекта или не приняли соглашение", "")
			fl = false
		}
	}
	return fl
}

// спасатель
func runSafely(bot *tgbotapi.BotAPI, chatID int64, name string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic in %s: %v", name, r)
			sendErrorMessage(bot, chatID, "Что-то пошло не так, попробуйте снова", "")
		}
	}()

	log.Printf("%s started now", name)
	fn()
}

// спасатель с лимитом единомоментно выполняющихся
func runSafelyLimited(
	bot *tgbotapi.BotAPI,
	chatID int64,
	name string,
	slots chan struct{},
	msgID int,
	fn func(),
) {
	waiting := false
	var globalMsgID int

	select {
	case slots <- struct{}{}:
		log.Printf(
			"%s started immediately; active=%d/%d; free=%d",
			name, len(slots), cap(slots), cap(slots)-len(slots),
		)
	default:
		waiting = true
		log.Printf("%s queued", name)
		msg := tgbotapi.NewMessage(chatID, "Данные приняты в очередь. Обработка начнётся, когда освободится место.")
		msg.ReplyParameters = &tgbotapi.ReplyParameters{MessageID: msgID}
		msgID, err := bot.Send(msg)
		if err != nil {
			log.Printf("error when send message: %v", err)
		}
		globalMsgID = msgID.MessageID
	}

	go func() {
		if waiting {
			slots <- struct{}{}
			delBot := tgbotapi.NewDeleteMessage(chatID, globalMsgID)
			_, err := bot.Request(delBot)
			if err != nil {
				log.Printf("error when del message: %v", err)
			}
			log.Printf(
				"%s started immediately; active=%d/%d; free=%d",
				name, len(slots), cap(slots), cap(slots)-len(slots),
			)
		}

		defer func() {
			<-slots
			log.Printf("%s reade: active=%d/%d; free=%d",
				name, len(slots), cap(slots), cap(slots)-len(slots),
			)
		}()
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic in %s: %v", name, r)
				sendErrorMessage(bot, chatID, "Что-то пошло не так, попробуйте снова", "")
			}
		}()

		fn()
	}()
}

func HandleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	userID := message.From.ID

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Intercepted panic: %v", r)
			sendErrorMessage(bot, chatID, "Что-то пошло не так, попробуйте снова", "")
			return
		}
	}()

	if message.Text != "" && message.Text == "/start" {
		deleteUserState(chatID)
		go showMainMenu(bot, chatID, userID)
		log.Printf("Активныч горутин: %d\n", runtime.NumGoroutine())
		return
	}

	if state, ok := getUserState(chatID); ok {
		for key := range state {
			switch key {
			case "search":
				if !userValid(bot, userID, chatID) {
					return
				}
				runSafelyLimited(bot, chatID, key, broSlot, message.MessageID, func() {
					handlerSearchMusic(bot, userID, chatID, message, "")
				})
				return
			case "download_music":
				if !userValid(bot, userID, chatID) {
					return
				}
				runSafely(bot, chatID, key, func() {
					handlerDownloadMusic(bot, userID, chatID, message)
				})
				return
			case "download_playlist":
				if !userValid(bot, userID, chatID) {
					return
				}
				// тут будет вызов обработчика
			default:
				log.Printf("неизвестная команда: %s", key)
				sendErrorMessage(bot, chatID, "Неизвестная команда", "menu")
				return
			}
		}
	}
}

func HandleCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
	chatID := callbackQuery.Message.Chat.ID
	data := callbackQuery.Data
	userID := callbackQuery.From.ID

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Перехвачена паника: %v", r)
			sendErrorMessage(bot, chatID, "Что-то пошло не так, попробуйте снова", "menu")
			return
		}
	}()

	switch data {
	case "show_menu":
		deleteUserState(chatID)
		sendSuccessMessage(bot, chatID, "Выберете действие", "command", 0)
		return
	case "search":
		setUserState(chatID, map[string]string{"search": ""})
		sendSuccessMessage(bot, chatID, "Введите название песни и исполнителя", "", 0)
		return
	case "download_music":
		setUserState(chatID, map[string]string{"download_music": ""})
		sendSuccessMessage(bot, chatID, "Отправьте аудио файл", "", 0)
		return
	case "download_playlist":
		setUserState(chatID, map[string]string{"download_playlist": ""})
		sendSuccessMessage(bot, chatID, "Отправьте файл плейлиста", "", 0)
		return
	case "accept":
		runSafely(bot, chatID, data, func() {
			AcceptHandler(bot, userID, chatID)
		})
		return
	case "check subscription":
		runSafely(bot, chatID, data, func() {
			ChekSub(bot, userID, chatID)
		})
	default:
		log.Printf("неизвестная команда: %s", data)
		sendErrorMessage(bot, chatID, "Неизвестная команда", "menu")
		return
	}
}
