package bot

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DataForParser struct {
	Title     string `json:"title"`
	Performer string `json:"performer"`
	UserID    int64  `json:"user_id"`
	ChatID    int64  `json:"chat_id"`
}

func handlerSearchMusic(bot *tgbotapi.BotAPI, userID, chatID int64, message *tgbotapi.Message) {
	var (
		title     string
		performer string
	)
	if strings.Contains(message.Text, " ") {
		parts := strings.Split(message.Text, " ")
		title = parts[0]
		performer = strings.Join(parts[1:], ",")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, file_id, err := mus_db.GetMusicByIndex(ctx, title, performer)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("error when searching music: %v", err)
		sendErrorMessage(bot, chatID, "Internal service error7: попробуйте позже", "menu")
		return
	} else if err == sql.ErrNoRows {
		data := &DataForParser{
			Title:     title,
			Performer: performer,
			UserID:    userID,
			ChatID:    chatID,
		}
		err := sendDataToServer(fmt.Sprintf("http://localhost:%s/home", conf.COMMUN_PORT), data)
		if err != nil {
			log.Printf("error when sending data to server: %v", err)
			sendErrorMessage(bot, chatID, "Internal service error8: попробуйте позже", "menu")
			return
		}
		log.Println("ВСЁ ЗАЕБИСЬ")
		return
	}
	sendFileAsAudio(bot, chatID, file_id, title, performer, "Вот ваш трек!")
}
func sendDataToServer(url string, data *DataForParser) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error serializing data: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func handlerDownloadMusic(bot *tgbotapi.BotAPI, userID, chatID int64, message *tgbotapi.Message) {
	log.Printf("прилёт %+v", message.Audio)
	if message.Audio != nil && (message.Audio.MimeType == "audio/mpeg" || message.Document.MimeType == "audio/mp4") {
		log.Println("условие прошли")
		file_id := message.Audio.FileID
		file_un_id := message.Audio.FileUniqueID
		performer := message.Audio.Performer
		title := message.Audio.Title
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := mus_db.InsertMusic(ctx, title, performer, file_id, file_un_id)
		if err != nil {
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
				sendSuccessMessage(bot, chatID, "Такой файл уже существует в базе, вы можете прослушивать файл отправленный вами", "menu")
				return
			} else {
				sendErrorMessage(bot, chatID, "Internal service error5: попробуйте позже", "menu")
				log.Printf("error whem insert track in db: %v", err)
				return
			}
		}
		log.Println("в базу залили")
		ctxUFIL, cancelUFIL := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelUFIL()
		if err := us_db.UpdateFileIDList(ctxUFIL, userID, file_id); err != nil {
			sendErrorMessage(bot, chatID, "Internal service error6: попробуйте позже", "menu")
			log.Printf("error whem update file id list: %v", err)
			return
		}
		sendFileAsAudio(bot, chatID, file_id, title, performer, "Ваш файл успешно сохранён, приятного прослушивания")
		return
	} else {
		sendErrorMessage(bot, chatID, "Такой формат файлов не поддерживаеться.", "menu")
		return
	}
}
