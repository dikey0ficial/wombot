package main

import (
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
)

type command struct {
	Name   string
	Is     func([]string, tg.Update) bool
	Action func([]string, tg.Update, User) error
}

var commands = []command{
	{
		Name: "start",
		Is: func(args []string, update tg.Update) bool {
			if strings.ToLower(args[0]) == "/start@"+bot.Self.UserName ||
				(!isGroup(update.Message) && isInList(args[0], []string{"/start", "/старт"})) {
				return true
			}
			return false
		},
		Action: func(args []string, update tg.Update, womb User) error {
			const longAnswer = "Доброе утро\n — Завести вомбата: `взять вомбата`\n — Помощь: https://telegra.ph/Pomoshch-10-28 (/help)\n — Канал бота, где есть нужная инфа: @wombatobot_channel\n Приятной игры!"
			if isGroup(update.Message) {
				_, err := replyToMsg(update.Message.MessageID, "Доброе утро! ((большинство комманд вомбота доступны только в лс))", update.Message.Chat.ID, bot)
				return err
			}
			_, err := replyToMsg(update.Message.MessageID, longAnswer, update.Message.Chat.ID, bot)
			return err
		},
	},
	{
		Name: "help",
		Is: func(args []string, update tg.Update) bool {
			if isInList(args[0], []string{"/help@" + bot.Self.UserName, "команды", "/хелп"}) ||
				(!isGroup(update.Message) && strings.ToLower(args[0]) == "/help") {
				return true
			}
			return false
		},
		Action: func(args []string, update tg.Update, womb User) error {
			_, err := replyToMsg(update.Message.MessageID, "https://telegra.ph/Pomoshch-10-28", update.Message.Chat.ID, bot)
			return err
		},
	},
	{
		Name: "about_bot",
		Is: func(args []string, update tg.Update) bool {
			if len(args) != 2 {
				return false
			} else if strings.ToLower(args[0]+" "+args[1]) == "о вомбате" {
				return true
			}
			return false
		},
		Action: func(args []string, update tg.Update, womb User) error {
			_, err := replyToMsgMD(update.Message.MessageID,
				"https://telegra.ph/O-vombote-10-29\n**если вы хотели узнать характеристики вомбата, используйте команду `о вомбате`**",
				update.Message.Chat.ID, bot,
			)
			return err
		},
	},
	{
		Name: "about_wombat",
		Is: func(args []string, update tg.Update) bool {
			if strings.HasPrefix(strings.Join(args, " "), "о вомбате") {
				return true
			}
			return false
		},
		Action: func(args []string, update tg.Update, womb User) error {
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			var strID string
			if len(args) == 3 {
				strID = strings.TrimSpace(strings.Join(args[2:], " "))
			} else if len(args) > 3 {
				_, err := replyToMsg(update.Message.MessageID, "Слишком много аргументов!", update.Message.Chat.ID, bot)
				if err != nil {
					return err
				}
			}
			var (
				tWomb User
			)
			if strID == "" {
				if update.Message.ReplyToMessage != nil {
					tWomb.ID = update.Message.ReplyToMessage.From.ID
					if c, err := users.CountDocuments(ctx, bson.M{"_id": tWomb.ID}); err != nil {
						return err
					} else if c == 0 {
						replyToMsg(update.Message.MessageID,
							"Данный пользователь не обладает вомбатом. (напищите свой ник, если хотите узнать о себе и с ответом)",
							update.Message.Chat.ID, bot,
						)
						return nil
					}
					if err := users.FindOne(ctx, bson.M{"_id": tWomb.ID}).Decode(&tWomb); err != nil {
						return err
					}
				} else if isInUsers {
					tWomb = womb
				} else {
					replyToMsg(update.Message.MessageID, "У вас нет вомбата", update.Message.Chat.ID, bot)
					return nil
				}
			} else if len([]rune(strID)) > 64 {
				replyToMsg(update.Message.MessageID, "Ошибка: слишком длинное имя", update.Message.Chat.ID, bot)
				return nil
			} else if !isValidName(strID) {
				replyToMsg(update.Message.MessageID, "Нелегальное имя!", update.Message.Chat.ID, bot)
				return nil
			} else if rCount, err :=
				users.CountDocuments(ctx, bson.M{"name": cins(strID)}); err == nil && rCount != 0 {
				err := users.FindOne(ctx, bson.M{"name": cins(strID)}).Decode(&tWomb)
				if err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else {
				replyToMsg(update.Message.MessageID, fmt.Sprintf("Ошибка: пользователя с именем %s не найдено", strID), update.Message.Chat.ID, bot)
				return nil
			}
			var clname string
			if c, err := clans.CountDocuments(ctx, bson.M{"members": tWomb.ID}); err != nil {
				return err
			} else if c != 0 {
				var uClan Clan
				if err := clans.FindOne(ctx, bson.M{"members": tWomb.ID}).Decode(&uClan); err != nil {
					return err
				}
				clname = "\\[" + uClan.Tag + "]"
			}
			strTitles := ""
			tCount := len(tWomb.Titles)
			if tCount != 0 {
				for _, id := range tWomb.Titles {
					rCount, err := titlesC.CountDocuments(ctx, bson.M{"_id": id})
					if err != nil {
						return err
					}
					if rCount == 0 {
						strTitles += fmt.Sprintf("Ошибка: титула с ID %d нет (ответьте командой /admin) |", id)
						continue
					}
					elem := Title{}
					err = titlesC.FindOne(ctx, bson.M{"_id": id}).Decode(&elem)
					if err != nil {
						return err
					}
					strTitles += fmt.Sprintf("%s (ID: %d) | ", elem.Name, id)
				}
				strTitles = strings.TrimSuffix(strTitles, " | ")
			} else {
				strTitles = "нет"
			}
			var sl string = "Не спит"
			if tWomb.Sleep {
				sl = "Спит"
			} else {
				sl = "Не спит"
			}
			abimg, err := getImgs(imgsC, "about")
			if err != nil {
				return err
			}
			_, err = replyWithPhotoMD(update.Message.MessageID, randImg(abimg), fmt.Sprintf(
				"Вомбат `%s` %s\nТитулы: %s\n 👁 %d XP\n ❤ %d здоровья\n ⚡ %d мощи\n 💰 %d шишей при себе\n 💤 %s",
				tWomb.Name, clname, strTitles, tWomb.XP, tWomb.Health, tWomb.Force, tWomb.Money, sl),
				update.Message.Chat.ID, bot,
			)
			return err
		},
	},
	{
		Name: "greeting_for_new_chat_members",
		Is: func(args []string, update tg.Update) bool {
			if update.Message.NewChatMembers != nil && len(update.Message.NewChatMembers) != 0 {
				return true
			}
			return false
		},
		Action: func(args []string, update tg.Update, womb User) error {
			isInUsers, err := getIsInUsers(update.Message.NewChatMembers[0].ID)
			if err != nil {
				return err
			}
			if update.Message.NewChatMembers[0].ID == bot.Self.ID {
				_, err = replyToMsg(
					update.Message.MessageID,
					fmt.Sprintf(randomString(
						"всем привет чат!1!1! /help@%s для инфы о коммандочках :з",
						"дарова вомбэты и вомбята. я ботяра. /help@%s -- инфа",
						"всем привет я бот /help@%s для подробностей",
						"короче, я бот с вомбатами. подробнее: /help@%s",
					), bot.Self.UserName),
					update.Message.Chat.ID,
					bot,
				)
			} else if isInUsers {
				_, err = replyToMsgMDNL(update.Message.MessageID,
					"Здравствуйте! Я [вомбот](t.me/wombatobot) — бот с вомбатами. "+
						"Рекомендую Вам завести вомбата, чтобы играть "+
						"вместе с другими участниками этого чата (^.^)",
					update.Message.Chat.ID, bot,
				)
			} else {
				_, err = replyToMsg(update.Message.MessageID, fmt.Sprintf("Добро пожаловать, вомбат `%s`!", womb.Name), update.Message.Chat.ID, bot)
			}
			return err
		},
	},
}
