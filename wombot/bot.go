package main

import (
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	*tg.BotAPI
}

type MessageOption func(*tg.MessageConfig) error

func MarkdownParseModeMessage(msg *tg.MessageConfig) error {
	msg.ParseMode = "markdown"
	return nil
}

func SetWebPagePreview(wpp bool) MessageOption {
	return func(msg *tg.MessageConfig) error {
		msg.DisableWebPagePreview = !wpp
		return nil
	}
}

func MessageSetReply(id int) MessageOption {
	return func(msg *tg.MessageConfig) error {
		msg.ReplyToMessageID = id
		return nil
	}
}

func (b Bot) SendMessage(message string, chatID int64, options ...MessageOption) (int, error) {
	msg := tg.NewMessage(chatID, message)

	for _, option := range options {
		if err := option(&msg); err != nil {
			return 0, err
		}
	}

	mess, err := b.Send(msg)

	return mess.MessageID, err
}

func (b Bot) ReplyWithMessage(messID int, message string, chatID int64, options ...MessageOption) (int, error) {
	return b.SendMessage(message, chatID, append(options, MessageSetReply(messID))...)
}

type PhotoOption func(*tg.PhotoConfig) error

func MarkdownParseModePhoto(msg *tg.PhotoConfig) error {
	msg.ParseMode = "markdown"
	return nil
}
func PhotoSetReply(id int) PhotoOption {
	return func(msg *tg.PhotoConfig) error {
		msg.ReplyToMessageID = id
		return nil
	}
}

func (b Bot) SendPhoto(id string, caption string, chatID int64, options ...PhotoOption) (int, error) {
	msg := tg.NewPhoto(chatID, tg.FileID(id))
	msg.Caption = caption
	for _, option := range options {
		if err := option(&msg); err != nil {
			return 0, err
		}
	}
	mess, err := b.Send(msg)
	return mess.MessageID, err
}

func (b Bot) ReplyWithPhoto(messID int, id, caption string, chatID int64, options ...PhotoOption) (int, error) {
	return b.SendPhoto(id, caption, chatID, append(options, PhotoSetReply(messID))...)
}

type EditMessageOption func(*tg.EditMessageTextConfig) error

func MarkdownParseModeEditText(msg *tg.EditMessageTextConfig) error {
	msg.ParseMode = "markdown"
	return nil
}

func SetWebPagePreviewEditText(wpp bool) EditMessageOption {
	return func(msg *tg.EditMessageTextConfig) error {
		msg.DisableWebPagePreview = !wpp
		return nil
	}
}

func (b Bot) EditMessage(messID int, newText string, chatID int64, options ...EditMessageOption) error {
	editConfig := tg.EditMessageTextConfig{
		BaseEdit: tg.BaseEdit{
			ChatID:    chatID,
			MessageID: messID,
		},
		Text: newText,
	}

	for _, option := range options {
		if err := option(&editConfig); err != nil {
			return err
		}
	}

	_, err := bot.Request(editConfig)
	return err
}

type EditCaptionOption func(*tg.EditMessageCaptionConfig) error

func MarkdownParseModeEditCaption(msg *tg.EditMessageCaptionConfig) error {
	msg.ParseMode = "markdown"
	return nil
}

func (b Bot) EditCaption(messID int, newText string, chatID int64, options ...EditCaptionOption) error {
	editConfig := tg.EditMessageCaptionConfig{
		BaseEdit: tg.BaseEdit{
			ChatID:    chatID,
			MessageID: messID,
		},
		Caption: newText,
	}

	for _, option := range options {
		if err := option(&editConfig); err != nil {
			return err
		}
	}

	_, err := bot.Request(editConfig)
	return err
}
