package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// newHelpButton создает кнопку с иконкой помощи, которая показывает диалог с информацией.
func newHelpButton(message string, parentWindow fyne.Window) *widget.Button {
	return widget.NewButtonWithIcon("", theme.HelpIcon(), func() {
		// Используем Label с включенным переносом строк для лучшего форматирования
		contentLabel := widget.NewLabel(message)
		contentLabel.Wrapping = fyne.TextWrapWord

		// Обернем в скроллер, если текст очень длинный
		scroller := container.NewVScroll(contentLabel)
		scroller.SetMinSize(fyne.Size{Width: 400, Height: 200}) // Задаем минимальный размер для содержимого диалога

		// Создаем кастомный диалог, чтобы можно было вставить скроллер
		// Кнопка "OK" уже есть по умолчанию в NewCustom
		customDialog := dialog.NewCustom("Help", "OK", scroller, parentWindow)
		customDialog.Resize(fyne.Size{Width: 500, Height: 300}) // Устанавливаем размер диалога
		customDialog.Show()
	})
}

// newFormFieldWithHelp создает FormItem с виджетом и кнопкой помощи справа.
// Изменено: fieldWidget теперь fyne.CanvasObject
func newFormFieldWithHelp(label string, fieldWidget fyne.CanvasObject, helpMessage string, parentWindow fyne.Window) *widget.FormItem {
	helpButton := newHelpButton(helpMessage, parentWindow)
	// Используем Border, чтобы разместить кнопку помощи справа от основного виджета
	fieldContainer := container.NewBorder(nil, nil, nil, helpButton, fieldWidget)
	return widget.NewFormItem(label, fieldContainer) // Исправлено: widget.NewFormItem
}

// newLabelWithHelp создает Label с кнопкой помощи справа.
func newLabelWithHelp(labelText string, textStyle fyne.TextStyle, helpMessage string, parentWindow fyne.Window) fyne.CanvasObject {
	label := widget.NewLabelWithStyle(labelText, fyne.TextAlignLeading, textStyle)
	helpButton := newHelpButton(helpMessage, parentWindow)
	return container.NewBorder(nil, nil, nil, helpButton, label) // Кнопка справа
}
