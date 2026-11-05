package telegram

import (
	"github.com/go-telegram/bot/models"
	"github.com/swokeqq/tatusashbot/internal/flow"
)

func keyboard(buttons []flow.Button) models.InlineKeyboardMarkup {
	rows := make([][]models.InlineKeyboardButton, 0, len(buttons))
	for _, button := range buttons {
		rows = append(rows, []models.InlineKeyboardButton{
			{Text: button.Text, CallbackData: button.Data},
		})
	}

	return models.InlineKeyboardMarkup{InlineKeyboard: rows}
}
