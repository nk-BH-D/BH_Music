package bot

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DataForParser struct {
	Title     string `json:"title"`
	Performer string `json:"performer"`
	ChatID    int64  `json:"chat_id"`
}

type DataFromParser struct {
	Title     string `json:"title"`
	Performer string `json:"performer"`
	URL       string `json:"url"`
	ChatID    int64  `json:"chat_id"`
}

// -
// обработчик для поиска музыки
// -
func handlerSearchMusic(bot *tgbotapi.BotAPI, userID, chatID int64, message *tgbotapi.Message, from_playlist string) {
	var (
		title     string
		performer string
	)
	if from_playlist != "" {
		parts := strings.Split(from_playlist, " ")
		title = parts[0]
		performer = strings.Join(parts[1:], ",")
	} else {
		if strings.Contains(message.Text, "\n") {
			parts := strings.Split(message.Text, "\n")
			title = parts[0]
			performer = parts[1]
		} else {
			log.Printf("data incorrect: %s", message.Text)
			sendErrorMessage(bot, chatID, "Данные не корректны", "menu")
			return
		}
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
			ChatID:    chatID,
		}
		file_un_id, fileID, err := sendDataToServer(bot, fmt.Sprintf("http://parser:%s/parse", conf.CAMOUFOX_PORT), data)
		if err != nil {
			log.Printf("error when sending data to server: %v", err)
			sendErrorMessage(bot, chatID, "Internal service error8: попробуйте позже", "menu")
			return
		}
		log.Println(fileID)
		// добавляем в базу
		ctxInsert, cancelInsert := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelInsert()
		if err := mus_db.InsertMusic(ctxInsert, title, performer, fileID, file_un_id); err != nil {
			log.Printf("error when inserting music: %v", err)
			sendErrorMessage(bot, chatID, "Internal service error9: попробуйте позже", "menu")
			return
		}
		ctxUFIL, cancelUFIL := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelUFIL()
		if err := us_db.UpdateFileIDList(ctxUFIL, userID, fileID); err != nil {
			log.Printf("error when updating user file id list: %v", err)
			sendErrorMessage(bot, chatID, "Internal service error10: попробуйте позже", "menu")
			return
		}
		return
	}
	sendFileAsAudio(bot, chatID, file_id, title, performer, "Вот ваш трек!")
}

func sendDataToServer(bot *tgbotapi.BotAPI, url string, data *DataForParser) (string, string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", "", fmt.Errorf("error serializing data: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		url,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", "", fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// отправляем запрос
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return "", "", fmt.Errorf("request timeout: %w", err)
		}
		return "", "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	} else {
		var respData DataFromParser
		if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
			return "", "", fmt.Errorf("error decoding response: %w", err)
		}
		log.Printf("Received from parser: %+v", respData)
		respData.URL = "https://ru-d3.drivemusic.me/dl/lOg3JY8ZCnK7oSnFCYLG0Q/1776925461/download_music/2024/11/arija-bespechnyjj-angel.mp3"
		file_un_id, fileID, err := downloadAndSendAudio(bot, respData.ChatID, respData.URL, respData.Title, respData.Performer)
		if err != nil {
			return "", "", fmt.Errorf("error downloading or sending audio: %w", err)
		}
		return file_un_id, fileID, nil
	}
}

func downloadAndSendAudio(bot *tgbotapi.BotAPI, chatID int64, fileURL, title, performer string) (string, string, error) {
	// cкачать файл
	resp, err := http.Get(fileURL)
	if err != nil {
		return "", "", fmt.Errorf("download error: %w", err)
	}
	defer resp.Body.Close()
	log.Println("скачали")

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("bad status when downloading: %d", resp.StatusCode)
	}

	//cоздать временный файл
	tmpFile, err := os.CreateTemp("", "*.mp3")
	if err != nil {
		return "", "", fmt.Errorf("temp file error: %w", err)
	}
	defer os.Remove(tmpFile.Name()) // удалить после
	defer tmpFile.Close()
	log.Printf("создали")

	const maxFileSize = 10 * 1024 * 1024 // 10 МБ
	limitedReader := io.LimitReader(resp.Body, maxFileSize+1)
	// cкопировать данные в файл
	writer, err := io.Copy(tmpFile, limitedReader) // ограничение 10 МБ
	if err != nil {
		return "", "", fmt.Errorf("write file error: %w", err)
	}
	log.Println("скопировали")
	if writer > maxFileSize {
		return "", "", fmt.Errorf("file too large: %d bytes", writer)
	}

	// отправить в Telegram
	audio := tgbotapi.NewAudio(chatID, tgbotapi.FilePath(tmpFile.Name()))
	audio.Caption = "Вот ваш трек!"
	audio.ReplyMarkup = createMenuKeyboard()
	audio.Title = title
	audio.Performer = performer
	msg, err := bot.Send(audio)
	if err != nil {
		return "", "", fmt.Errorf("telegram send error: %w", err)
	}

	// получить file_id
	if msg.Audio == nil {
		return "", "", fmt.Errorf("no audio in response")
	}
	log.Println("ВСЁ ЗАЕБИСЬ")
	return msg.Audio.FileUniqueID, msg.Audio.FileID, nil
}

// -
// -

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
