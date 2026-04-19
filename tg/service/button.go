package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

//func createRatingKeeydoard() tgbotapi.InlineKeyboardMarkup {
//	return tgbotapi.NewInlineKeyboardMarkup(
//		tgbotapi.NewInlineKeyboardRow(
//			tgbotapi.NewInlineKeyboardButtonData("👍", "good"),
//			tgbotapi.NewInlineKeyboardButtonData("👎", "bad"),
//		),
//	)
//}

func createCommandKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Найти музыку", "search"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Загрузить музыку", "download_music"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Загрузить плейлис", "download_playlist"),
		),
	)
}

func createMenuKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Меню", "show_menu"),
		),
	)
}
