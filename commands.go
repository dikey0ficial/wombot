package main

import (
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
	"strconv"
	"strings"
	"time"
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
	{
		Name: "support",
		Is: func(args []string, update tg.Update) bool {
			txt := strings.ToLower(strings.Join(args, " "))
			if isPrefixInList(txt, []string{"одмен!", "/admin@" + bot.Self.UserName, "/support@" + bot.Self.UserName, "/bug@" + bot.Self.UserName}) {
				return true
			} else if !isGroup(update.Message) && isPrefixInList(txt, []string{"/admin", "/support", "/bug", "/админ", "/сап", "/саппорт", "/баг"}) {
				return true
			}
			return false
		},
		Action: func(args []string, update tg.Update, womb User) error {
			var (
				isGr           = "из чата "
				isInUsers, err = getIsInUsers(update.Message.From.ID)
				txt            = strings.ToLower(strings.Join(args, " "))
			)
			if err != nil {
				return err
			}
			if isGroup(update.Message) {
				isGr = "из группы "
			}
			if len(args) < 2 {
				if update.Message.ReplyToMessage == nil {
					replyToMsg(update.Message.MessageID, "Ты чаво... где письмо??", update.Message.Chat.ID, bot)
					return nil
				}
				r := update.Message.ReplyToMessage
				_, serr := sendMsg(
					fmt.Sprintf(
						"%d %d \nписьмо %s(%d @%s) от %d (@%s isInUsers: %v) (mt: %s bt: %s), отвечающее на: \n%s\n(id:%d fr:%d @%s) (mt:%s, bt: %s)",
						update.Message.MessageID, update.Message.Chat.ID, isGr, update.Message.Chat.ID, update.Message.Chat.UserName,
						update.Message.From.ID, update.Message.From.UserName, isInUsers,
						time.Unix(int64(update.Message.Date), 0).String(), time.Now().String(),
						r.Text, r.MessageID, r.From.ID, r.From.UserName,
						time.Unix(int64(r.Date), 0).String(), time.Now().String(),
					),
					conf.SupChatID, bot,
				)
				_, err = replyToMsg(update.Message.MessageID, "Письмо отправлено! Скоро (или нет) придёт ответ", update.Message.Chat.ID, bot)
				if err != nil {
					if serr != nil {
						return fmt.Errorf("Two errors: %v and %v", serr, err)
					}
					return err
				}
			} else {
				if update.Message.ReplyToMessage == nil {
					msg := strings.Join(args[1:], " ")
					_, serr := sendMsg(
						fmt.Sprintf(
							"%d %d \nписьмо %s%d (@%s) от %d (@%s isInUsers: %v): \n%s\n(mt: %s bt:%s)",
							update.Message.MessageID, update.Message.Chat.ID, isGr, update.Message.Chat.ID, update.Message.Chat.UserName, update.Message.From.ID,
							update.Message.From.UserName, isInUsers, msg,
							time.Unix(int64(update.Message.Date), 0).String(), time.Now().String(),
						),
						conf.SupChatID, bot,
					)
					_, err := replyToMsg(update.Message.MessageID, "Письмо отправлено! Скоро (или нет) придёт ответ", update.Message.Chat.ID, bot)
					if err != nil {
						if serr != nil {
							return fmt.Errorf("Two errors: %v and %v", serr, err)
						}
						return err
					}
				} else {
					r := update.Message.ReplyToMessage
					_, serr := sendMsg(
						fmt.Sprintf(
							"%d %d \nписьмо %s(%d @%s) от %d (@%s isInUsers: %v), отвечающее на: \n%s\n(id:%d fr:%d @%s) (mt: %s bt: %s) с текстом:\n%s\n(mt: %s bt: %s)",
							update.Message.MessageID, update.Message.Chat.ID, isGr, update.Message.Chat.ID, update.Message.Chat.UserName,
							update.Message.From.ID, update.Message.From.UserName,
							isInUsers, r.Text, r.MessageID, r.From.ID, r.From.UserName,
							time.Unix(int64(update.Message.Date), 0).String(), time.Now().String(),
							txt,
							time.Unix(int64(r.Date), 0).String(), time.Now().String(),
						), conf.SupChatID, bot,
					)
					_, err := replyToMsg(update.Message.MessageID, "Письмо отправлено! Скоро (или нет) придёт ответ", update.Message.Chat.ID, bot)
					if err != nil {
						if serr != nil {
							return fmt.Errorf("Two errors: %v and %v", serr, err)
						}
						return err
					}
				}
			}
			return nil
		},
	},
	{
		Name: "take_wombat",
		Is: func(args []string, update tg.Update) bool {
			if isInList(strings.ToLower(strings.Join(args, " ")),
				[]string{"взять вомбата", "купить вомбата у арабов", "хочу вомбата"},
			) {
				return true
			}
			return false
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if isGroup(update.Message) {
				_, err := replyToMsg(update.Message.MessageID, "данная команда работает (мб только пока) только в лс)", update.Message.Chat.ID, bot)
				return err
			}
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if isInUsers {
				_, err := replyToMsg(update.Message.MessageID,
					"У тебя как бы уже есть вомбат лолкек. Если хочешь от него избавиться, то напиши `приготовить шашлык`",
					update.Message.Chat.ID, bot,
				)
				return err
			}
			newWomb := User{
				ID:     update.Message.From.ID,
				Name:   "Вомбат_" + strconv.Itoa(int(update.Message.From.ID)),
				XP:     0,
				Health: 5,
				Force:  2,
				Money:  10,
				Titles: []uint16{},
				Sleep:  false,
			}
			_, err = users.InsertOne(ctx, &newWomb)
			if err != nil {
				return err
			}
			newimg, err := getImgs(imgsC, "new")
			if err != nil {
				return err
			}
			_, err = replyWithPhoto(update.Message.MessageID,
				randImg(newimg), fmt.Sprintf(
					"Поздравляю, у тебя появился вомбат! Ему выдалось имя `%s`. Ты можешь поменять имя командой `Поменять имя [имя]` за 3 монеты",
					newWomb.Name),
				update.Message.Chat.ID, bot,
			)
			return err
		},
	},
	{
		Name: "schweine",
		Is: func(args []string, update tg.Update) bool {
			if strings.HasPrefix(strings.ToLower(args[0]), "хрю") {
				return true
			}
			return false
		},
		Action: func(args []string, update tg.Update, womb User) error {
			schweineImgs, err := getImgs(imgsC, "schweine")
			if err != nil {
				return err
			}
			_, err = replyWithPhoto(update.Message.MessageID,
				randImg(schweineImgs),
				"АХТУНГ ШВАЙНЕ УИИИИИИИИИИИИИИИИИИИИИИИИИ",
				update.Message.Chat.ID, bot,
			)
			return err
		},
	},
	{
		Name: "delete_wombat",
		Is: func(args []string, update tg.Update) bool {
			return isInList(
				strings.ToLower(strings.Join(args, " ")),
				[]string{
					"приготовить шашлык", "продать вомбата арабам",
					"слить вомбата в унитаз", "расстрелять вомбата",
				},
			)
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if isGroup(update.Message) {
				_, err := replyToMsg(update.Message.MessageID, "данная команда работает (мб только пока) только в лс)", update.Message.Chat.ID, bot)
				return err
			}
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if !isInUsers {
				_, err := replyToMsg(update.Message.MessageID, "Но у вас нет вомбата...", update.Message.Chat.ID, bot)
				return err
			}
			if hasTitle(1, womb.Titles) {
				_, err := replyToMsg(update.Message.MessageID,
					"Ошибка: вы лишены права уничтожать вомбата; ответьте на это сообщение командой /admin для объяснений",
					update.Message.Chat.ID, bot,
				)
				return err
			}
			_, err = users.DeleteOne(ctx, wombFilter(womb))
			if err != nil {
				return err
			}
			kill, err := getImgs(imgsC, "kill")
			if err != nil {
				return err
			}
			_, err = replyWithPhoto(update.Message.MessageID,
				randImg(kill), "Вы уничтожили вомбата в количестве 1 штука. Вы - нехорошее существо", update.Message.Chat.ID,
				bot,
			)
			return err
		},
	},
}
