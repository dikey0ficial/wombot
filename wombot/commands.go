package main

import (
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"math"
	"math/rand"
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
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					fmt.Sprintf(randomString(
						"всем привет чат!1!1! /help@%s для инфы о коммандочках :з",
						"дарова вомбэты и вомбята. я ботяра. /help@%s -- инфа",
						"всем привет я бот /help@%s для подробностей",
						"короче, я бот с вомбатами. подробнее: /help@%s",
					), bot.Self.UserName),
					update.Message.Chat.ID,
				)
			} else if isInUsers {
				_, err = bot.ReplyWithMessage(update.Message.MessageID,
					"Здравствуйте! Я [вомбот](t.me/wombatobot) — бот с вомбатами. "+
						"Рекомендую Вам завести вомбата, чтобы играть "+
						"вместе с другими участниками этого чата."+
						"подробнее: /help@wombatobot",
					update.Message.Chat.ID, MarkdownParseModeMessage, SetWebPagePreview(false),
				)
			} else {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, fmt.Sprintf("Добро пожаловать, вомбат `%s`!", womb.Name), update.Message.Chat.ID)
			}
			return err
		},
	},
	{
		Name: "bad_update_check",
		Is: func(args []string, update tg.Update) bool {
			return update.Message == nil || update.Message.Chat == nil || update.Message.From == nil || args == nil || len(args) == 0
		},
		Action: func([]string, tg.Update, User) error {
			return nil
		},
	},
	{
		Name: "start",
		Is: func(args []string, update tg.Update) bool {
			return (!isGroup(update.Message) && isInList(strings.ToLower(args[0]), []string{"/start", "/старт"})) || strings.ToLower(args[0]) == "/start@"+bot.Self.UserName
		},
		Action: func(args []string, update tg.Update, womb User) error {
			const longAnswer = "Доброе утро\n — Завести вомбата: `взять вомбата`\n — Помощь: https://telegra.ph/Pomoshch-10-28 (/help)\n — Канал бота, где есть нужная инфа: @wombatobot_channel\n Приятной игры!"
			if isGroup(update.Message) {
				_, err := bot.ReplyWithMessage(update.Message.MessageID, "Доброе утро! ((большинство комманд вомбота доступны только в лс))", update.Message.Chat.ID)
				return err
			}
			_, err := bot.ReplyWithMessage(update.Message.MessageID, longAnswer, update.Message.Chat.ID)
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
			_, err := bot.ReplyWithMessage(update.Message.MessageID, "https://telegra.ph/Pomoshch-10-28", update.Message.Chat.ID)
			return err
		},
	},
	{
		Name: "about_bot",
		Is: func(args []string, update tg.Update) bool {
			if len(args) != 2 {
				return false
			} else if strings.ToLower(args[0]+" "+args[1]) == "о вомботе" {
				return true
			}
			return false
		},
		Action: func(args []string, update tg.Update, womb User) error {
			_, err := bot.ReplyWithMessage(update.Message.MessageID,
				"https://telegra.ph/O-vombote-10-29\n**если вы хотели узнать характеристики вомбата, используйте команду `о вомбате`**",
				update.Message.Chat.ID, MarkdownParseModeMessage,
			)
			return err
		},
	},
	{
		Name: "about_wombat",
		Is: func(args []string, update tg.Update) bool {
			if strings.HasPrefix(strings.ToLower(strings.Join(args, " ")), "о вомбате") {
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
				_, err := bot.ReplyWithMessage(update.Message.MessageID, "Слишком много аргументов!", update.Message.Chat.ID)
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
						bot.ReplyWithMessage(update.Message.MessageID,
							"Данный пользователь не обладает вомбатом. (напищите свой ник, если хотите узнать о себе и с ответом)",
							update.Message.Chat.ID,
						)
						return nil
					}
					if err := users.FindOne(ctx, bson.M{"_id": tWomb.ID}).Decode(&tWomb); err != nil {
						return err
					}
				} else if isInUsers {
					tWomb = womb
				} else {
					bot.ReplyWithMessage(update.Message.MessageID, "У вас нет вомбата", update.Message.Chat.ID)
					return nil
				}
			} else if len([]rune(strID)) > 64 {
				bot.ReplyWithMessage(update.Message.MessageID, "Ошибка: слишком длинное имя", update.Message.Chat.ID)
				return nil
			} else if !isValidName(strID) {
				bot.ReplyWithMessage(update.Message.MessageID, "Нелегальное имя!", update.Message.Chat.ID)
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
				bot.ReplyWithMessage(update.Message.MessageID, fmt.Sprintf("Ошибка: пользователя с именем %s не найдено", strID), update.Message.Chat.ID)
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
			_, err = bot.ReplyWithPhoto(update.Message.MessageID, randImg(abimg), fmt.Sprintf(
				"Вомбат [%s](tg://user?id=%d) %s\nТитулы: %s\n 👁 %d XP\n ❤ %d здоровья\n ⚡ %d мощи\n 💰 %d шишей при себе\n 💤 %s",
				tWomb.Name, tWomb.ID, clname, strTitles, tWomb.XP, tWomb.Health, tWomb.Force, tWomb.Money, sl),
				update.Message.Chat.ID, MarkdownParseModePhoto,
			)
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
					bot.ReplyWithMessage(update.Message.MessageID, "Ты чаво... где письмо??", update.Message.Chat.ID)
					return nil
				}
				r := update.Message.ReplyToMessage
				_, serr := bot.SendMessage(
					fmt.Sprintf(
						"%d %d \nписьмо %s(%d @%s) от %d (@%s isInUsers: %v) (mt: %s bt: %s), отвечающее на: \n%s\n(id:%d fr:%d @%s) (mt:%s, bt: %s)",
						update.Message.MessageID, update.Message.Chat.ID, isGr, update.Message.Chat.ID, update.Message.Chat.UserName,
						update.Message.From.ID, update.Message.From.UserName, isInUsers,
						time.Unix(int64(update.Message.Date), 0).String(), time.Now().String(),
						r.Text, r.MessageID, r.From.ID, r.From.UserName,
						time.Unix(int64(r.Date), 0).String(), time.Now().String(),
					),
					conf.SupChatID,
				)
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Письмо отправлено! Скоро (или нет) придёт ответ", update.Message.Chat.ID)
				if err != nil {
					if serr != nil {
						return fmt.Errorf("Two errors: %v and %v", serr, err)
					}
					return err
				}
			} else {
				if update.Message.ReplyToMessage == nil {
					msg := strings.Join(args[1:], " ")
					_, serr := bot.SendMessage(
						fmt.Sprintf(
							"%d %d \nписьмо %s%d (@%s) от %d (@%s isInUsers: %v): \n%s\n(mt: %s bt:%s)",
							update.Message.MessageID, update.Message.Chat.ID, isGr, update.Message.Chat.ID, update.Message.Chat.UserName, update.Message.From.ID,
							update.Message.From.UserName, isInUsers, msg,
							time.Unix(int64(update.Message.Date), 0).String(), time.Now().String(),
						),
						conf.SupChatID,
					)
					_, err := bot.ReplyWithMessage(update.Message.MessageID, "Письмо отправлено! Скоро (или нет) придёт ответ", update.Message.Chat.ID)
					if err != nil {
						if serr != nil {
							return fmt.Errorf("Two errors: %v and %v", serr, err)
						}
						return err
					}
				} else {
					r := update.Message.ReplyToMessage
					_, serr := bot.SendMessage(
						fmt.Sprintf(
							"%d %d \nписьмо %s(%d @%s) от %d (@%s isInUsers: %v), отвечающее на: \n%s\n(id:%d fr:%d @%s) (mt: %s bt: %s) с текстом:\n%s\n(mt: %s bt: %s)",
							update.Message.MessageID, update.Message.Chat.ID, isGr, update.Message.Chat.ID, update.Message.Chat.UserName,
							update.Message.From.ID, update.Message.From.UserName,
							isInUsers, r.Text, r.MessageID, r.From.ID, r.From.UserName,
							time.Unix(int64(update.Message.Date), 0).String(), time.Now().String(),
							txt,
							time.Unix(int64(r.Date), 0).String(), time.Now().String(),
						), conf.SupChatID,
					)
					_, err := bot.ReplyWithMessage(update.Message.MessageID, "Письмо отправлено! Скоро (или нет) придёт ответ", update.Message.Chat.ID)
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
			return isInList(
				strings.ToLower(strings.Join(args, " ")),
				[]string{"взять вомбата", "купить вомбата у арабов", "хочу вомбата"},
			)
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if isGroup(update.Message) {
				_, err := bot.ReplyWithMessage(update.Message.MessageID, "данная команда работает (мб только пока) только в лс)", update.Message.Chat.ID)
				return err
			}
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if isInUsers {
				_, err := bot.ReplyWithMessage(update.Message.MessageID,
					"У тебя как бы уже есть вомбат лолкек. Если хочешь от него избавиться, то напиши `приготовить шашлык`",
					update.Message.Chat.ID,
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
			iiuCache.Put(update.Message.From.ID, true)
			newimg, err := getImgs(imgsC, "new")
			if err != nil {
				return err
			}
			_, err = bot.ReplyWithPhoto(update.Message.MessageID,
				randImg(newimg), fmt.Sprintf(
					"Поздравляю, у тебя появился вомбат! Ему выдалось имя `%s`. Ты можешь поменять имя командой `Поменять имя [имя]` за 3 монеты",
					newWomb.Name),
				update.Message.Chat.ID,
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
			_, err = bot.ReplyWithPhoto(update.Message.MessageID,
				randImg(schweineImgs),
				"АХТУНГ ШВАЙНЕ УИИИИИИИИИИИИИИИИИИИИИИИИИ",
				update.Message.Chat.ID,
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
				_, err := bot.ReplyWithMessage(update.Message.MessageID, "данная команда работает (мб только пока) только в лс)", update.Message.Chat.ID)
				return err
			}
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if !isInUsers {
				_, err := bot.ReplyWithMessage(update.Message.MessageID, "Но у вас нет вомбата...", update.Message.Chat.ID)
				return err
			}
			if hasTitle(1, womb.Titles) {
				_, err := bot.ReplyWithMessage(
					update.Message.MessageID,
					"Ошибка: вы лишены права уничтожать вомбата; ответьте на это сообщение командой /admin для объяснений",
					update.Message.Chat.ID,
				)
				return err
			}
			if c, err := clans.CountDocuments(ctx, bson.M{"members": update.Message.From.ID}); err != nil {
				return err
			} else if c != 0 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Вы состоите в клане; выйдите перед этим из клана",
					update.Message.Chat.ID,
				)
				return err
			}
			_, err = users.DeleteOne(ctx, wombFilter(womb))
			if err != nil {
				return err
			}
			iiuCache.Put(update.Message.From.ID, false)
			kill, err := getImgs(imgsC, "kill")
			if err != nil {
				return err
			}
			_, err = bot.ReplyWithPhoto(update.Message.MessageID,
				randImg(kill), "Вы уничтожили вомбата в количестве 1 штука. Вы - нехорошее существо", update.Message.Chat.ID,
			)
			return err
		},
	},
	{
		Name: "change_name",
		Is: func(args []string, update tg.Update) bool {
			return strings.HasPrefix(
				strings.ToLower(strings.Join(args, " ")),
				"поменять имя",
			)
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if isGroup(update.Message) {
				_, err := bot.ReplyWithMessage(update.Message.MessageID, "данная команда работает (мб только пока) только в лс)", update.Message.Chat.ID)
				return err
			}
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if !isInUsers {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Да блин нафиг, вы вобмата забыли завести!!!!!!!", update.Message.From.ID)
				return err
			} else if len(args) != 3 {
				if len(args) == 2 {
					_, err = bot.ReplyWithMessage(update.Message.MessageID, "вомбату нужно имя! ты его не указал", update.Message.From.ID)
				} else {
					_, err = bot.ReplyWithMessage(update.Message.MessageID, "слишком много аргументов...", update.Message.From.ID)
				}
				return err
			} else if hasTitle(1, womb.Titles) {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Тебе нельзя, ты спамер (оспорить: /admin)", update.Message.From.ID)
				return err
			} else if womb.Money < 3 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Мало шишей блин нафиг!!!!", update.Message.From.ID)
				return err
			}
			name := args[2]
			if womb.Name == name {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "зачем", update.Message.From.ID)
				return err
			} else if len([]rune(name)) > 64 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Слишком длинный никнейм!", update.Message.From.ID)
				return err
			} else if isInList(name, []string{"вoмбoт", "вoмбoт", "вомбoт", "вомбот", "бот", "bot", "бoт", "bоt",
				"авто", "auto"}) {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Такие никнеймы заводить нельзя", update.Message.From.ID)
				return err
			} else if !isValidName(name) {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Нелегальное имя:(\n", update.Message.From.ID)
				return err
			}
			rCount, err := users.CountDocuments(ctx, bson.M{"name": cins(name)})
			if err != nil {
				return err
			} else if rCount != 0 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, fmt.Sprintf("Никнейм `%s` уже занят(", name), update.Message.From.ID)
				return err
			}
			womb.Money -= 3
			caseName := strings.Join(args[2:], " ")
			womb.Name = caseName
			err = docUpd(womb, bson.M{"_id": update.Message.From.ID}, users)
			if err != nil {
				return err
			}
			_, err = bot.ReplyWithMessage(update.Message.MessageID,
				fmt.Sprintf("Теперь вашего вомбата зовут %s. С вашего счёта сняли 3 шиша", caseName),
				update.Message.From.ID,
			)
			return err
		},
	},
	{
		Name: "find_money",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(strings.Join(args, " ")) == "поиск денег"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if !isInUsers {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "А ты куда? У тебя вомбата нет...", update.Message.Chat.ID)
				return err
			}

			if womb.Money < 1 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Охранники тебя прогнали; они требуют шиш за проход, а у тебя ни шиша нет", update.Message.Chat.ID)
				return err
			}
			womb.Money--
			rand.Seed(time.Now().UnixNano())
			if ch := rand.Int(); ch%2 == 0 || hasTitle(2, womb.Titles) && (ch%2 == 0 || ch%3 == 0) {
				rand.Seed(time.Now().UnixNano())
				win := rand.Intn(9) + 1
				womb.Money += uint32(win)
				if addXP := rand.Intn(512 - 1); addXP < 5 {
					womb.XP += uint32(addXP)
					_, err = bot.ReplyWithMessage(update.Message.MessageID,
						fmt.Sprintf(
							"Поздравляем! Вы нашли на дороге %d шишей, а ещё вам дали %d XP! Теперь у вас %d шишей при себе и %d XP",
							win, addXP, womb.Money, womb.XP,
						),
						update.Message.Chat.ID,
					)
				} else {
					_, err = bot.ReplyWithMessage(update.Message.MessageID,
						fmt.Sprintf(
							"Поздравляем! Вы нашли на дороге %d шишей! Теперь их у вас при себе %d", win, womb.Money,
						),
						update.Message.Chat.ID,
					)
				}
				if err != nil {
					return err
				}
			} else {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID, "Вы заплатили один шиш охранникам денежной дорожки, но увы, вы так ничего и не нашли",
					update.Message.Chat.ID,
				)
				if err != nil {
					return err
				}
			}
			return docUpd(womb, wombFilter(womb), users)
		},
	},
	{
		Name: "shop",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(strings.Join(args, " ")) == "магазин"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			_, err := bot.ReplyWithMessage(
				update.Message.MessageID,
				strings.Join(
					[]string{
						"Магазин:",
						fmt.Sprintf(" — 1 здоровье — %d ш", 5+womb.XP%100),
						fmt.Sprintf(" — 1 мощь — %d ш", 3+womb.XP%100),
						" — Квес — 256 ш",
						" — Вадшам — 250'000 ш",
						"Для покупки использовать команду 'купить [название_объекта] ([кол-во])",
					},
					"\n",
				),
				update.Message.Chat.ID,
			)
			return err
		},
	},
	{
		Name: "buy",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[0]) == "купить"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if len(args) == 1 {
				_, err := bot.ReplyWithMessage(update.Message.MessageID, "купить", update.Message.Chat.ID)
				return err
			}
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if !isInUsers {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "у тебя недостаточно вомбатов чтобы купить (нужен минимум один)", update.Message.Chat.ID)
				return err
			}
			switch strings.ToLower(args[1]) {
			case "здоровья":
				fallthrough
			case "здоровье":
				if len(args) > 3 {
					_, err := bot.ReplyWithMessage(update.Message.MessageID, "Ошибка: слишком много аргументов...", update.Message.Chat.ID)
					return err
				}
				var amount uint32 = 1
				if len(args) == 3 {
					if val, err := strconv.ParseUint(args[2], 10, 32); err == nil {
						if val == 0 {
							_, err = bot.ReplyWithMessage(update.Message.MessageID, "Поздравляю! Теперь у вас одна шиза и ещё одна шиза", update.Message.Chat.ID)
							return err
						}
						amount = uint32(val)
					} else {
						_, err = bot.ReplyWithMessage(update.Message.MessageID, "Ошибка: число должно быть неотрицательным, целым и меньше 2^32", update.Message.Chat.ID)
						return err
					}
				}

				var costOfOne uint32 = 5 + womb.XP%100

				if womb.Money < uint32(amount)*costOfOne {
					_, err = bot.ReplyWithMessage(
						update.Message.MessageID,
						fmt.Sprintf(
							"Надо накопить побольше шишей! 1 здоровье = %d шишей",
							costOfOne,
						),
						update.Message.Chat.ID,
					)
					return err
				}
				if uint64(womb.Health+amount) > uint64(math.Pow(2, 32)) {
					_, err = bot.ReplyWithMessage(update.Message.MessageID,
						"Ошибка: вы достигли максимального количества здоровья (2 в 32 степени). Если это вас возмущает, ответьте командой /admin",
						update.Message.Chat.ID,
					)
					return err
				}
				womb.Money -= uint32(amount) * costOfOne
				womb.Health += amount
				err = docUpd(womb, wombFilter(womb), users)
				if err != nil {
					return err
				}
				_, err = bot.ReplyWithMessage(update.Message.MessageID,
					fmt.Sprintf("Поздравляю! Теперь у вас %d здоровья и %d шишей при себе", womb.Health, womb.Money),
					update.Message.Chat.ID,
				)
				return err
			case "силу":
				fallthrough
			case "сила":
				fallthrough
			case "силы":
				fallthrough
			case "мощи":
				fallthrough
			case "мощь":
				if len(args) > 3 {
					_, err := bot.ReplyWithMessage(update.Message.MessageID, "Ошибка: слишком много аргументов...", update.Message.Chat.ID)
					return err
				}
				var amount uint32 = 1
				if len(args) == 3 {
					if val, err := strconv.ParseUint(args[2], 10, 32); err == nil {
						if val == 0 {
							_, err = bot.ReplyWithMessage(update.Message.MessageID, "Поздравляю! Теперь у вас одна шиза и ещё одна шиза", update.Message.Chat.ID)
							return err
						}
						amount = uint32(val)
					} else {
						_, err = bot.ReplyWithMessage(update.Message.MessageID, "Ошибка: число должно быть неотрицательным, целым и меньше 2^32", update.Message.Chat.ID)
						return err
					}
				}

				var costOfOne uint32 = 3 + womb.XP%100

				if womb.Money < uint32(amount)*costOfOne {
					_, err = bot.ReplyWithMessage(
						update.Message.MessageID,
						fmt.Sprintf(
							"Надо накопить побольше шишей! 1 мощь = %d шиша",
							costOfOne,
						),
						update.Message.Chat.ID,
					)
					return err
				}
				if uint64(womb.Force+amount) > uint64(math.Pow(2, 32)) {
					_, err = bot.ReplyWithMessage(update.Message.MessageID,
						"Ошибка: вы достигли максимального количества мощи (2 в 32 степени). Если это вас возмущает, ответьте командой /admin",
						update.Message.Chat.ID,
					)
					return err
				}
				womb.Money -= uint32(amount) * costOfOne
				womb.Force += amount
				err = docUpd(womb, wombFilter(womb), users)
				if err != nil {
					return err
				}
				_, err = bot.ReplyWithMessage(update.Message.MessageID,
					fmt.Sprintf("Поздравляю! Теперь у вас %d силы и %d шишей при себе", womb.Force, womb.Money),
					update.Message.Chat.ID,
				)
				return err
			case "вадшамка":
				fallthrough
			case "вадшама":
				fallthrough
			case "вадшамку":
				fallthrough
			case "вадшамки":
				fallthrough
			case "вадшам":
				if len(args) != 2 {
					_, err = bot.ReplyWithMessage(update.Message.MessageID, "ужас !! слишком много аргументов!!!", update.Message.Chat.ID)
					return err
				} else if hasTitle(4, womb.Titles) {
					_, err = bot.ReplyWithMessage(update.Message.MessageID, "у вас уже есть вадшам", update.Message.Chat.ID)
					return err
				} else if womb.Money < 250005 {
					_, err = bot.ReplyWithMessage(update.Message.MessageID, "Ошибка: недостаточно шишей для покупки (требуется 250000 + 5)", update.Message.Chat.ID)
					return err
				}
				womb.Money -= 250000
				womb.Titles = append(womb.Titles, 4)
				err = docUpd(womb, wombFilter(womb), users)
				if err != nil {
					return err
				}
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Теперь вы вадшамообладатель", update.Message.Chat.ID)
			case "квес":
				fallthrough
			case "квеса":
				fallthrough
			case "квесу":
				fallthrough
			case "qwess":
				if len(args) != 2 {
					_, err = bot.ReplyWithMessage(update.Message.MessageID, "Слишком много аргументов!", update.Message.Chat.ID)
					return err
				} else if womb.Money < 256 {
					leps, err := getImgs(imgsC, "leps")
					if err != nil {
						return err
					}
					_, err = bot.ReplyWithPhoto(update.Message.MessageID,
						randImg(leps),
						"Вы подошли к ближайшему ларьку, но, увы, кролик-Лепс на кассе сказал, что надо 256 шишей, а у вас, к сожалению, меньше",
						update.Message.Chat.ID,
					)
					return err
				}
				qwess, err := getImgs(imgsC, "qwess")
				if err != nil {
					return err
				}
				if !(hasTitle(2, womb.Titles)) {
					womb.Titles = append(womb.Titles, 2)
					womb.Money -= 256
					err = docUpd(womb, wombFilter(womb), users)
					if err != nil {
						return err
					}
					_, err = bot.ReplyWithPhoto(update.Message.MessageID,
						randImg(qwess),
						"Вы купили чудесного вкуса квес у кролика-Лепса в ларьке за 256 шишей. Глотнув этот напиток, вы поняли, что получили новый титул с ID 2",
						update.Message.Chat.ID,
					)
				} else {
					womb.Money -= 256
					err = docUpd(womb, wombFilter(womb), users)
					if err != nil {
						return err
					}
					if err != nil {
						return err
					}
					_, err = bot.ReplyWithPhoto(update.Message.MessageID,
						randImg(qwess),
						"Вы вновь купили вкусного квеса у того же кролика-Лепса в том же ларьке за 256 шишей. \"Он так освежает, я чувствую себя человеком\" — думаете вы. Ах, как вкусён квес!",
						update.Message.Chat.ID,
					)
					return err
				}
			default:
				_, err = bot.ReplyWithMessage(update.Message.MessageID, fmt.Sprintf("Что такое %s?", args[1]), update.Message.Chat.ID)
				return err
			}
			return nil
		},
	},
	{
		Name: "about_title",
		Is: func(args []string, update tg.Update) bool {
			return strings.HasPrefix(strings.ToLower(strings.Join(args, " ")), "о титуле")
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if len(args) < 3 {
				_, err := bot.ReplyWithMessage(update.Message.MessageID, "Ошибка: пустой ID титула", update.Message.Chat.ID)
				return err
			}
			strID := strings.Join(args[2:], " ")
			i, err := strconv.ParseInt(strID, 10, 64)
			if err != nil {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Ошибка: неправильный синтаксис. Синтаксис команды: `о титуле {ID титула}`", update.Message.Chat.ID)
				return err
			} else {
			}
			ID := uint16(i)
			rCount, err := titlesC.CountDocuments(ctx, bson.M{"_id": ID})
			if err != nil {
				return err
			}
			if rCount == 0 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, fmt.Sprintf("Ошибка: не найдено титула по ID %d", ID), update.Message.Chat.ID)
				return err
			}
			elem := Title{}
			err = titlesC.FindOne(ctx, bson.M{"_id": ID}).Decode(&elem)
			if err != nil {
				return err
			}
			_, err = bot.ReplyWithMessage(update.Message.MessageID, fmt.Sprintf("%s | ID: %d\n%s", elem.Name, ID, elem.Desc), update.Message.Chat.ID)
			return err
		},
	},
	{
		Name: "sleep",
		Is: func(args []string, update tg.Update) bool {
			return isInList(strings.ToLower(strings.Join(args, " ")), []string{"лечь спать", "споке", "спать", "споть"})
		},
		Action: func(args []string, update tg.Update, womb User) error {
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if !isInUsers {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "У тебя нет вомбата, иди спи сам", update.Message.Chat.ID)
				return err
			} else if womb.Sleep {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Твой вомбат уже спит. Если хочешь проснуться, то напиши `проснуться` (логика)", update.Message.Chat.ID)
				return err
			}
			womb.Sleep = true
			err = docUpd(womb, wombFilter(womb), users)
			if err != nil {
				return err
			}
			sleep, err := getImgs(imgsC, "sleep")
			if err != nil {
				return err
			}
			_, err = bot.ReplyWithPhoto(update.Message.MessageID, randImg(sleep), "Вы легли спать. Спокойного сна!", update.Message.Chat.ID)
			return err
		},
	},
	{
		Name: "unsleep",
		Is: func(args []string, update tg.Update) bool {
			return isInList(strings.ToLower(strings.Join(args, " ")), []string{"добрутро", "проснуться", "не спать", "не споть", "рота подъём"})
		},
		Action: func(args []string, update tg.Update, womb User) error {
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if !isInUsers {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "У тебя нет вомбата, буди себя сам", update.Message.Chat.ID)
				return err
			} else if !womb.Sleep {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Твой вомбат и так не спит, может ты хотел лечь спать? (команда `лечь спать` (опять логика))",
					update.Message.Chat.ID,
				)
				return err
			}
			womb.Sleep = false
			var msg string = "Вомбат проснулся без каких-либо проишествий"
			rand.Seed(time.Now().UnixNano())
			if rand.Intn(2) == 1 {
				switch rand.Intn(9) {
				case 0:
					i := uint32(rand.Intn(15) + 1)
					womb.Health += i
					msg = fmt.Sprintf("Вомбат отлично выспался. Офигенный сон ему дал %d здоровья", i)
				case 1:
					i := uint32(rand.Intn(10) + 1)
					womb.Force += i
					msg = fmt.Sprintf("Встав, вомбат почувствовал силу в своих лапах! +%d мощи", i)
				case 3:
					i := uint32(rand.Intn(100) + 1)
					womb.Money += i
					msg = fmt.Sprintf("Проснувшись, вомбат увидел мешок, содержащий %d шишей. Кто бы мог его оставить?", i)
				case 4:
					if womb.Money > 50 {
						womb.Money = 50
					} else if womb.Money > 10 {
						womb.Money = 10
					} else {
						break
					}
					msg = fmt.Sprintf("Ужас!!! Вас обокрали!!! У вас теперь только %d шишей при себе!", womb.Money)
				case 5:
					if womb.Health <= 5 {
						break
					}
					womb.Health--
					msg = "Шатаясь, вомбат встал с кровати. Он себя чувствует ужасно. -1 здоровья"
				case 6:
					if womb.Force <= 2 {
						break
					}
					womb.Force--
					msg = "Ваш вомбат чувствует слабость... -1 мощи"
				case 7:
					msg = "Ваш вомбат встал и загадочно улыбнулся..."
				case 8:
					i := uint32(rand.Intn(4) + 1)
					womb.XP += i
					msg = fmt.Sprintf("Ваш вомбат увидел странный сон. Почесав подбородок, он получает %d XP", i)
				}
			}
			err = docUpd(womb, wombFilter(womb), users)
			if err != nil {
				return err
			}
			unsleep, err := getImgs(imgsC, "unsleep")
			if err != nil {
				return err
			}
			_, err = bot.ReplyWithPhoto(update.Message.MessageID, randImg(unsleep), msg, update.Message.Chat.ID)
			return err
		},
	},
	{
		Name: "send_shishes",
		Is: func(args []string, update tg.Update) bool {
			return strings.HasPrefix(strings.ToLower(strings.Join(args, " ")), "перевести шиши")
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if len(args) == 2 {
				_, err := bot.ReplyWithMessage(update.Message.MessageID, "так и запишем", update.Message.Chat.ID)
				return err
			}
			cargs := args[2:]
			if len(cargs) < 2 {
				_, err := bot.ReplyWithMessage(update.Message.MessageID,
					"Ошибка: вы пропустили аргумент(ы). Синтаксис команды: `перевести шиши [кол-во] [никнейм получателя]`",
					update.Message.Chat.ID,
				)
				return err
			} else if len(cargs) > 2 {
				_, err := bot.ReplyWithMessage(update.Message.MessageID,
					"Ошибка: слишком много аргументов. Синтаксис команды: `перевести шиши [кол-во] [никнейм получателя]`",
					update.Message.Chat.ID,
				)
				return err
			}
			var (
				amount uint64
				err    error
			)
			if amount, err = strconv.ParseUint(cargs[0], 10, 32); err != nil {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"нелегальные у Вас какие-то числа",
					update.Message.Chat.ID,
				)
				return err
			}
			var ID int64
			name := cargs[1]
			if len([]rune(name)) > 64 {
				_, err := bot.ReplyWithMessage(update.Message.MessageID, "Слишком длинный никнейм", update.Message.Chat.ID)
				return err
			} else if !isValidName(name) {
				_, err := bot.ReplyWithMessage(update.Message.MessageID, "Нелегальное имя", update.Message.Chat.ID)
				return err
			} else if rCount, err := users.CountDocuments(
				ctx, bson.M{"name": cins(name)}); err != nil {
				return err
			} else if rCount == 0 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, fmt.Sprintf("Ошибка: вомбата с именем %s не найдено", name), update.Message.Chat.ID)
				return err
			}
			var tWomb User
			err = users.FindOne(ctx, bson.M{"name": cins(name)}).Decode(&tWomb)
			if err != nil {
				return err
			}
			ID = tWomb.ID
			if uint64(womb.Money) < amount {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID, "Ошибка: на Вашем счету шишей меньше указанного количества",
					update.Message.Chat.ID,
				)
				return err
			}
			if amount == 0 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID,
					"Ошибка: количество переводимых шишей должно быть больше нуля",
					update.Message.Chat.ID,
				)
				return err
			}
			if ID == update.Message.From.ID {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Ты читер блин нафиг!!!!!! нидам тебе самому себе перевести", update.Message.Chat.ID)
				return err
			}
			rCount, err := users.CountDocuments(ctx, bson.M{"_id": ID})
			if err != nil {
				return err
			}
			if rCount == 0 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID,
					fmt.Sprintf("Ошибка: пользователя с ID %d не найдено", ID),
					update.Message.Chat.ID,
				)
				return err
			}
			womb.Money -= uint32(amount)
			tWomb.Money += uint32(amount)
			err = docUpd(tWomb, bson.M{"_id": ID}, users)
			if err != nil {
				return err
			}
			err = docUpd(womb, wombFilter(womb), users)
			if err != nil {
				return err
			}
			_, err = bot.ReplyWithMessage(update.Message.MessageID,
				fmt.Sprintf("Вы успешно перевели %d шишей на счёт %s. Теперь у вас %d шишей при себе",
					amount, tWomb.Name, womb.Money), update.Message.Chat.ID,
			)
			if err != nil {
				return err
			}
			_, err = bot.SendMessage(fmt.Sprintf("Пользователь %s перевёл вам %d шишей. Теперь у вас %d шишей при себе",
				womb.Name, amount, tWomb.Money), ID,
			)
			return err
		},
	},
	{
		Name: "rating",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[0]) == "рейтинг" || strings.ToLower(args[0]) == "топ"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			var (
				name  string = "xp"
				queue int8   = -1
			)
			if len(args) >= 2 && len(args) < 4 {
				if isInList(args[1], []string{"шиши", "деньги", "money"}) {
					name = "money"
				} else if isInList(args[1], []string{"хп", "опыт", "xp", "хрю"}) {
					name = "xp"
				} else if isInList(args[1], []string{"здоровье", "хил", "хеалтх", "health"}) {
					name = "health"
				} else if isInList(args[1], []string{"сила", "мощь", "force", "мощъ"}) {
					name = "force"
				} else {
					_, err := bot.ReplyWithMessage(update.Message.MessageID, fmt.Sprintf("не понимаю, что значит %s", args[1]), update.Message.Chat.ID)
					return err
				}
				if len(args) == 3 {
					if isInList(args[2], []string{"+", "плюс", "++", "увеличение"}) {
						queue = 1
					} else if isInList(args[2], []string{"-", "минус", "--", "уменьшение"}) {
						queue = -1
					} else {
						_, err := bot.ReplyWithMessage(update.Message.MessageID, fmt.Sprintf("не понимаю, что значит %s", args[2]), update.Message.Chat.ID)
						return err
					}
				}
			} else if len(args) != 1 {
				_, err := bot.ReplyWithMessage(update.Message.MessageID, "слишком много аргументов", update.Message.Chat.ID)
				return err
			}
			opts := options.Find()
			opts.SetSort(bson.M{name: queue})
			opts.SetLimit(10)
			cur, err := users.Find(ctx, bson.M{}, opts)
			if err != nil {
				return err
			}
			var rating []User
			for cur.Next(ctx) {
				var w User
				cur.Decode(&w)
				rating = append(rating, w)
			}
			var msg string = "Топ-10 вомбатов по "
			switch name {
			case "money":
				msg += "шишам "
			case "xp":
				msg += "XP "
			case "health":
				msg += "здоровью "
			case "force":
				msg += "мощи "
			}
			msg += "в порядке "
			if queue == 1 {
				msg += "увеличения:"
			} else if queue == -1 {
				msg += "уменьшения:"
			} else {
				return err
			}
			msg += "\n"
			for num, w := range rating {
				switch name {
				case "money":
					msg += fmt.Sprintf("%d | [%s](tg://user?id=%d) | %d шишей при себе\n", num+1, w.Name, w.ID, w.Money)
				case "xp":
					msg += fmt.Sprintf("%d | [%s](tg://user?id=%d) | %d XP\n", num+1, w.Name, w.ID, w.XP)
				case "health":
					msg += fmt.Sprintf("%d | [%s](tg://user?id=%d) | %d здоровья\n", num+1, w.Name, w.ID, w.Health)
				case "force":
					msg += fmt.Sprintf("%d | [%s](tg://user?id=%d) | %d мощи\n", num+1, w.Name, w.ID, w.Force)
				}
			}
			msg = strings.TrimSuffix(msg, "\n")
			_, err = bot.ReplyWithMessage(update.Message.MessageID, msg, update.Message.Chat.ID, MarkdownParseModeMessage)
			return err
		},
	},
	// laughter commands
	{
		Name: "want_to_laugh",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(update.Message.Text) == "хочу ржать"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}

			if !isInUsers {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"у Вас нет вомбата, а без вомбата не пустят на ржекич",
					update.Message.Chat.ID,
				)
				return err
			}

			if !isGroup(update.Message) {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"ржение доступно только в групповых чатах",
					update.Message.Chat.ID,
				)
				return err
			}

			if c, err := laughters.CountDocuments(ctx, bson.M{"members": update.Message.From.ID}); err != nil {
				return err
			} else if c != 0 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Вы уже учавствуете в ржекиче в этом или другом чате",
					update.Message.Chat.ID,
				)
				return err
			}

			if c, err := laughters.CountDocuments(ctx, bson.M{"_id": update.Message.Chat.ID}); err != nil {
				return err
			} else if c == 0 {
				_, err = laughters.InsertOne(
					ctx,
					bson.M{
						"_id":     update.Message.Chat.ID,
						"active":  true,
						"leader":  update.Message.From.ID,
						"members": []int64{update.Message.From.ID},
					},
				)
				if err != nil {
					return err
				}
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"вы первый пожелавший ржать, тем самым вы стали лидером сегоднешнего ржанья! оно не начнётся без вашей команды `начать ржение`",
					update.Message.Chat.ID,
				)
				if err != nil {
					return err
				}
			} else {
				_, err = laughters.UpdateOne(
					ctx,
					bson.M{
						"_id": update.Message.Chat.ID,
					},
					bson.M{
						"$push": bson.M{
							"members": update.Message.From.ID,
						},
					},
				)
				if err != nil {
					return err
				}
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"вы приняты в собрание ржения!!!",
					update.Message.Chat.ID,
				)
				if err != nil {
					return err
				}
			}

			return nil
		},
	},
	{
		Name: "want_not_to_laugh",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(update.Message.Text) == "не хочу ржать"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if c, err := laughters.CountDocuments(ctx, bson.M{"members": update.Message.From.ID}); err != nil {
				return err
			} else if c == 0 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Вы и так не учавствуете ни в одном ржении",
					update.Message.Chat.ID,
				)
				return err
			}
			var (
				nLghter Laughter
				newLead int64 = nLghter.Leader
			)
			err := laughters.FindOne(ctx, bson.M{"members": update.Message.From.ID}).Decode(&nLghter)
			if nLghter.Leader == update.Message.From.ID {
				for _, i := range nLghter.Members {
					if i != update.Message.From.ID {
						newLead = i
					}
				}
			}
			_, err = laughters.UpdateOne(
				ctx,
				bson.M{"members": update.Message.From.ID},
				bson.M{
					"$pull": bson.M{
						"members": update.Message.From.ID,
					},
					"$set": bson.M{
						"leader": newLead,
					},
				},
			)
			if err != nil {
				return err
			}

			_, err = bot.ReplyWithMessage(
				update.Message.MessageID,
				"несмотря на печаль в своих глазах, вы вышли из ржанного собрания.",
				update.Message.Chat.ID,
			)
			return err
		},
	},
	{
		Name: "laughter_status",
		Is: func(args []string, update tg.Update) bool {
			if len(args) < 2 {
				return false
			}
			return strings.ToLower(strings.Join(args[:2], " ")) == "статус ржения"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}

			var nLghter Laughter

			switch len(args) {
			case 2:
				if !isInUsers {
					_, err = bot.ReplyWithMessage(
						update.Message.MessageID,
						"у вас нет вомбата, чтобы посмотреть его статус ржения. "+
							"добавьте никнейм другого вомбата или `чат` к команде, чтобы узнать статус вомбата или этого чата",
						update.Message.Chat.ID,
					)
					return err
				}
				if c, err := laughters.CountDocuments(ctx, bson.M{"members": update.Message.From.ID}); err != nil {
					return err
				} else if c == 0 {
					_, err = bot.ReplyWithMessage(
						update.Message.MessageID,
						"Вы не участвуете ни в одном ржении",
						update.Message.Chat.ID,
					)
					return err
				}

				err = laughters.FindOne(ctx, bson.M{"members": update.Message.From.ID}).Decode(&nLghter)

				if err != nil {
					return err
				}
			case 3:
				if args[2] == "чат" {
					if !isGroup(update.Message) {
						_, err = bot.ReplyWithMessage(
							update.Message.MessageID,
							"ржение бывает только в групповых чатах",
							update.Message.Chat.ID,
						)
						return err
					}

					if c, err := laughters.CountDocuments(ctx, bson.M{"_id": update.Message.Chat.ID}); err != nil {
						return err
					} else if c == 0 {
						_, err = bot.ReplyWithMessage(
							update.Message.MessageID,
							"В чате нет ни одного активного ржения",
							update.Message.Chat.ID,
						)
						return err
					}

					err = laughters.FindOne(ctx, bson.M{"_id": update.Message.Chat.ID}).Decode(&nLghter)

					if err != nil {
						return err
					}

					if !nLghter.Active {
						_, err = bot.ReplyWithMessage(
							update.Message.MessageID,
							"В чате нет ни одного активного ржения",
							update.Message.Chat.ID,
						)
						return err
					}
				} else {
					if c, err := users.CountDocuments(ctx, bson.M{"name": cins(args[2])}); err != nil {
						return err
					} else if c == 0 {
						_, err = bot.ReplyWithMessage(
							update.Message.MessageID,
							"Вомбата с таким именем не найдено",
							update.Message.Chat.ID,
						)
						return err
					}

					var tWomb User

					err = users.FindOne(ctx, bson.M{"name": cins(args[2])}).Decode(&tWomb)
					if err != nil {
						return err
					}

					if c, err := laughters.CountDocuments(ctx, bson.M{"members": tWomb.ID}); err != nil {
						return err
					} else if c == 0 {
						_, err = bot.ReplyWithMessage(
							update.Message.MessageID,
							"Вомбат "+tWomb.Name+" не участвуете ни в одном ржении",
							update.Message.Chat.ID,
						)
						return err
					}

					err = laughters.FindOne(ctx, bson.M{"members": tWomb.ID}).Decode(&nLghter)
				}
			default:
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"чёт многовато аргументов",
					update.Message.Chat.ID,
				)
				return err
			}

			var builder strings.Builder

			builder.WriteString("ℹ Ржение:\n")

			var wombs = make([]User, 0)

			for _, memb := range nLghter.Members {
				var tWomb User
				err = users.FindOne(ctx, bson.M{"_id": memb}).Decode(&tWomb)
				if err != nil {
					continue
				}
				wombs = append(wombs, tWomb)
			}

			builder.WriteString("  Участники ржения:\n")

			for _, tWomb := range wombs {
				builder.WriteString(
					fmt.Sprintf("   - [%s](tg://user?id=%d)", tWomb.Name, tWomb.ID),
				)
				if tWomb.ID == nLghter.Leader {
					builder.WriteString(" (Лидер)")
				}
				builder.WriteRune('\n')
			}
			builder.WriteRune('\n')

			if e := time.Now().Sub(nLghter.LastStartTime); e < 24*time.Hour {
				left := (24 * time.Hour) - e
				builder.WriteString(
					fmt.Sprintf(
						"До следующей возможности запустить ржение осталось %d часов %d минут",
						int64(left.Hours()), int64(left.Minutes())-int64(left.Hours())*60,
					),
				)
			}

			_, err = bot.ReplyWithMessage(
				update.Message.MessageID,
				builder.String(),
				update.Message.Chat.ID,
				MarkdownParseModeMessage,
			)

			return nil
		},
	},
	// subcommand handlers
	{
		Name: "attack",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[0]) == "атака"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if len(args) == 1 {
				_, err := bot.ReplyWithMessage(update.Message.MessageID, "неправда", update.Message.Chat.ID)
				return err
			}
			for _, cmd := range attackCommands {
				if cmd.Is(args, update) {
					err := cmd.Action(args, update, womb)
					if err != nil {
						err = fmt.Errorf("%s: %v", cmd.Name, err)
					}
					return err
				}
			}
			_, err := bot.ReplyWithMessage(update.Message.MessageID, "не знаю такой команды", update.Message.Chat.ID)
			return err
		},
	},
	{
		Name: "wombank",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[0]) == "вомбанк"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if len(args) == 1 {
				_, err := bot.ReplyWithMessage(update.Message.MessageID, "неправда", update.Message.Chat.ID)
				return err
			}
			for _, cmd := range bankCommands {
				if cmd.Is(args, update) {
					err := cmd.Action(args, update, womb)
					if err != nil {
						err = fmt.Errorf("%s: %v", cmd.Name, err)
					}
					return err
				}
			}
			_, err := bot.ReplyWithMessage(update.Message.MessageID, "не знаю такой команды", update.Message.Chat.ID)
			return err
		},
	},
	{
		Name: "clans",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[0]) == "клан"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if len(args) == 1 {
				_, err := bot.ReplyWithMessage(update.Message.MessageID, "угадал", update.Message.Chat.ID)
				return err
			}
			for _, cmd := range clanCommands {
				if cmd.Is(args, update) {
					err := cmd.Action(args, update, womb)
					if err != nil {
						err = fmt.Errorf("%s: %v", cmd.Name, err)
					}
					return err
				}
			}
			_, err := bot.ReplyWithMessage(update.Message.MessageID, "не знаю такой команды, чесслово", update.Message.Chat.ID)
			return err
		},
	},
	{
		Name: "devtools",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[0]) == "devtools"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if len(args) == 1 {
				return nil
			} else if womb.Titles == nil || !hasTitle(0, womb.Titles) {
				return nil
			}
			for _, cmd := range devtoolsCommands {
				if cmd.Is(args, update) {
					err := cmd.Action(args, update, womb)
					if err != nil {
						err = fmt.Errorf("%s: %v", cmd.Name, err)
					}
					return err
				}
			}
			return nil
		},
	},
	// service commands
	{
		Name: "send_msg",
		Is: func(args []string, update tg.Update) bool {
			s := strings.ToLower(args[0])
			return s == "bot.SendMessage" || s == "send_msg"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if !hasTitle(3, womb.Titles) {
				return nil
			} else if len(args) < 3 {
				_, err := bot.ReplyWithMessage(update.Message.MessageID, "мало аргументов", update.Message.Chat.ID)
				return err
			}
			to, err := strconv.Atoi(args[1])
			if err != nil {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "error converting string to int64", update.Message.Chat.ID)
				return err
			}
			_, err = bot.SendMessage(strings.Join(args[2:], " "), int64(to), MarkdownParseModeMessage)
			if err != nil {
				return err
			}
			_, err = bot.ReplyWithMessage(update.Message.MessageID, "Запрос отправлен успешно!", update.Message.Chat.ID)
			return err
		},
	},
	{
		Name: "reply_to_msg",
		Is: func(args []string, update tg.Update) bool {
			s := strings.ToLower(args[0])
			return s == "bot.ReplyWithMessage" || s == "reply_to_msg"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			debl.Println("!!!")
			if !hasTitle(3, womb.Titles) {
				debl.Println("!!!")
				return nil
			} else if len(args) < 4 {
				_, err := bot.ReplyWithMessage(update.Message.MessageID, "мало аргументов", update.Message.Chat.ID)
				return err
			}
			sto, err := strconv.Atoi(args[1])
			if err != nil {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "error converting #1 string to int64", update.Message.Chat.ID)
				return err
			}
			rto, err := strconv.Atoi(args[2])
			if err != nil {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "error converting #2 string to int64", update.Message.Chat.ID)
				return err
			}
			_, err = bot.ReplyWithMessage(rto, strings.Join(args[3:], " "), int64(sto), MarkdownParseModeMessage)
			if err != nil {
				return err
			}
			_, err = bot.ReplyWithMessage(update.Message.MessageID, "успешно!", update.Message.Chat.ID)
			return err
		},
	},
	{
		Name: "send_photo",
		Is: func(args []string, update tg.Update) bool {
			s := strings.ToLower(args[0])
			return s == "bot.SendPhoto" || s == "send_photo"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if !hasTitle(3, womb.Titles) {
				return nil
			} else if len(args) < 2 {
				_, err := bot.ReplyWithMessage(update.Message.MessageID, "мало аргументов", update.Message.Chat.ID)
				return err
			}
			_, err := bot.ReplyWithPhoto(update.Message.MessageID, args[1], "", update.Message.Chat.ID)
			return err
		},
	},
	{
		Name: "photo_id",
		Is: func(args []string, update tg.Update) bool {
			s := strings.ToLower(args[0])
			return s == "photoid" || s == "photo_id"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if !hasTitle(3, womb.Titles) {
				return nil
			} else if len(update.Message.Photo) == 0 {
				_, err := bot.ReplyWithMessage(update.Message.MessageID, "нет фотографий", update.Message.Chat.ID)
				return err
			}
			var msg string
			for _, img := range update.Message.Photo {
				msg += "`" + img.FileID + "`\n"
			}
			_, err := bot.ReplyWithMessage(update.Message.MessageID, msg, update.Message.Chat.ID, MarkdownParseModeMessage)
			return err
		},
	},
	// support chat processor
	{
		Name: "support_chat_checker",
		Is: func(args []string, update tg.Update) bool {
			return update.Message.Chat.ID == conf.SupChatID && update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.From.ID == bot.Self.ID
		},
		Action: func(args []string, update tg.Update, womb User) error {
			strMessID := strings.Fields(update.Message.ReplyToMessage.Text)[0]
			omID, err := strconv.ParseInt(strMessID, 10, 64)
			if err != nil {
				return err
			}
			strPeer := strings.Fields(update.Message.ReplyToMessage.Text)[1]
			peer, err := strconv.ParseInt(strPeer, 10, 64)
			if err != nil {
				return err
			}
			if update.Message.From.UserName != "" {
				_, err = bot.ReplyWithMessage(
					int(omID),
					fmt.Sprintf(
						"Ответ от [админа](t.me/%s): \n%s",
						update.Message.From.UserName,
						update.Message.Text,
					),
					peer,
					MarkdownParseModeMessage, SetWebPagePreview(false),
				)
			} else {
				_, err = bot.ReplyWithMessage(
					int(omID),
					fmt.Sprintf(
						"Ответ от админа (для обращений: %d): \n%s",
						update.Message.From.ID,
						update.Message.Text,
					),
					peer,
					MarkdownParseModeMessage,
				)
			}
			return err
		},
	},
}

var attackCommands = []command{
	{
		Name: "attack",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "атака"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			_, err := bot.ReplyWithMessage(update.Message.MessageID, strings.Repeat("атака ", 42), update.Message.Chat.ID)
			return err
		},
	},
	{
		Name: "status",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "статус"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			var ID int64
			if len(args) == 2 {
				if !isInUsers {
					_, err = bot.ReplyWithMessage(update.Message.MessageID, "Но у вас вомбата нет...", update.Message.Chat.ID)
					return err
				}
				ID = int64(update.Message.From.ID)
			} else if len(args) > 3 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Атака статус: слишком много аргументов", update.Message.Chat.ID)
				return err
			} else {
				strID := args[2]
				if rCount, err := users.CountDocuments(ctx,
					bson.M{"name": cins(strID)}); err != nil {
					return err
				} else if rCount == 0 {
					_, err = bot.ReplyWithMessage(update.Message.MessageID, fmt.Sprintf("Пользователя с никнеймом `%s` не найдено", strID), update.Message.Chat.ID)
					return err
				}
				var tWomb User
				err = users.FindOne(ctx, bson.M{"name": cins(strID)}).Decode(&tWomb)
				if err != nil {
					return err
				}
				ID = tWomb.ID
			}
			var at Attack
			if is, isFrom := isInAttacks(ID, attacks); isFrom {
				a, err := getAttackByWomb(ID, true, attacks)
				if err != nil {
					return err
				}
				at = a
			} else if is {
				a, err := getAttackByWomb(update.Message.From.ID, false, attacks)
				if err != nil {
					return err
				}
				at = a
			} else {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "У этого вомбата атак нет", update.Message.Chat.ID)
				return err
			}
			var fromWomb, toWomb User
			err = users.FindOne(ctx, bson.M{"_id": at.From}).Decode(&fromWomb)
			if err != nil {
				return err
			}
			err = users.FindOne(ctx, bson.M{"_id": at.To}).Decode(&toWomb)
			if err != nil {
				return err
			}
			_, err = bot.ReplyWithMessage(
				update.Message.MessageID,
				fmt.Sprintf(
					"От: [%s](tg://user?id=%d)\nКому: [%s](tg://user?id=%d)\n",
					fromWomb.Name, fromWomb.ID,
					toWomb.Name, toWomb.ID,
				),
				update.Message.Chat.ID,
				MarkdownParseModeMessage,
			)
			return err
		},
	},
	{
		Name: "to",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "на"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if len(args) < 3 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Атака на: на кого?", update.Message.Chat.ID)
				return err
			} else if len(args) > 3 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Атака на: слишком много аргументов", update.Message.Chat.ID)
				return err
			} else if !isInUsers {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Вы не можете атаковать в виду остутствия вомбата", update.Message.Chat.ID)
				return err
			} else if womb.Sleep {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Но вы же спите...", update.Message.Chat.ID)
				return err
			}
			strID := args[2]
			var (
				ID    int64
				tWomb User
			)
			if is, isFrom := isInAttacks(update.Message.From.ID, attacks); isFrom {
				at, err := getAttackByWomb(update.Message.From.ID, true, attacks)
				if err != nil && err != errNoAttack {
					return err
				}
				var aWomb User
				err = users.FindOne(ctx, bson.M{"_id": at.To}).Decode(&aWomb)
				if err != nil {
					return err
				}
				_, err = bot.ReplyWithMessage(update.Message.MessageID,
					fmt.Sprintf(
						"Вы уже атакуете вомбата `%s`. Чтобы отозвать атаку, напишите `атака отмена`",
						aWomb.Name,
					),
					update.Message.Chat.ID,
					MarkdownParseModeMessage,
				)
				return err
			} else if is {
				at, err := getAttackByWomb(update.Message.From.ID, false, attacks)
				if err != nil && err != errNoAttack {
					return err
				}
				var aWomb User
				err = users.FindOne(ctx, bson.M{"_id": at.From}).Decode(&aWomb)
				if err != nil {
					return err
				}
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					fmt.Sprintf(
						"Вас уже атакует вомбат `%s`. Чтобы отклонить атаку, напишите `атака отмена`",
						aWomb.Name,
					),
					update.Message.Chat.ID,
					MarkdownParseModeMessage,
				)
				return err
			}
			if rCount, err := users.CountDocuments(ctx,
				bson.M{"name": cins(strID)}); err != nil && rCount != 0 {
				return err
			} else if rCount == 0 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					fmt.Sprintf(
						"Пользователя с именем `%s` не найдено",
						strID),
					update.Message.Chat.ID,
				)
				return err
			}
			err = users.FindOne(ctx, bson.M{"name": cins(strID)}).Decode(&tWomb)
			if err != nil {
				return err
			}
			ID = tWomb.ID
			if ID == int64(update.Message.MessageID) {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "„Главная борьба в нашей жизни — борьба с самим собой“ (c) какой-то философ", update.Message.From.ID)
				return err
			}
			err = users.FindOne(ctx, bson.M{"_id": ID}).Decode(&tWomb)
			if err != nil {
				return err
			}
			if tWomb.ID == womb.ID {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "„Главная борьба в нашей жизни — борьба с самим собой“ (c) какой-то философ", update.Message.From.ID)
				return err
			} else if tWomb.Sleep {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					fmt.Sprintf(
						"Вомбат %s спит. Его атаковать не получится",
						tWomb.Name,
					),
					update.Message.Chat.ID,
				)
				return err
			} else if is, isFrom := isInAttacks(ID, attacks); isFrom {
				at, err := getAttackByWomb(ID, true, attacks)
				if err != nil && err != errNoAttack {
					return err
				}
				var aWomb User
				err = users.FindOne(ctx, bson.M{"_id": at.To}).Decode(&aWomb)
				if err != nil {
					return err
				}
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID, fmt.Sprintf(
						"%s уже атакует вомбата %s. Попросите %s решить данную проблему",
						strID, aWomb.Name, strID,
					),
					update.Message.Chat.ID,
					MarkdownParseModeMessage,
				)
				return err
			} else if is {
				at, err := getAttackByWomb(int64(update.Message.MessageID), false, attacks)
				if err != nil && err != errNoAttack {
					return err
				}
				var aWomb User
				err = users.FindOne(ctx, bson.M{"_id": at.From}).Decode(&aWomb)
				if err != nil {
					return err
				}
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					fmt.Sprintf(
						"Вомбат %s уже атакуется %s. Попросите %s решить данную проблему",
						strID, aWomb.Name, strID,
					),
					update.Message.Chat.ID,
				)
				return err
			}
			var newAt = Attack{
				ID:   strconv.Itoa(int(update.Message.From.ID)) + "_" + strconv.Itoa(int(ID)),
				From: int64(update.Message.From.ID),
				To:   ID,
			}
			_, err = attacks.InsertOne(ctx, newAt)
			if err != nil {
				return err
			}
			_, err = bot.ReplyWithMessage(
				update.Message.MessageID,
				fmt.Sprintf(
					"Вы отправили вомбата атаковать %s. Ждём ответа!\nОтменить можно командой `атака отмена`",
					tWomb.Name,
				),
				update.Message.Chat.ID,
			)
			if err != nil {
				return err
			}
			_, err = bot.SendMessage(
				fmt.Sprintf(
					"Ужас! Вас атакует %s. Предпримите какие-нибудь меры: отмените атаку (`атака отмена`) или примите (`атака принять`)",
					womb.Name,
				),
				tWomb.ID,
			)
			return err
		},
	},
	{
		Name: "cancel",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "отмена"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if len(args) > 2 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "атака отмена: слишком много аргументов", update.Message.Chat.ID)
				return err
			} else if !isInUsers {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "какая атака, у тебя вобмата нет", update.Message.Chat.ID)
				return err
			}
			var at Attack
			if is, isFrom := isInAttacks(update.Message.From.ID, attacks); isFrom {
				a, err := getAttackByWomb(update.Message.From.ID, true, attacks)
				if err != nil {
					return err
				}
				at = a
			} else if is {
				a, err := getAttackByWomb(update.Message.From.ID, false, attacks)
				if err != nil {
					return err
				}
				at = a
			} else {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Атаки с вами не найдено...", update.Message.Chat.ID)
				return err
			}
			_, err = attacks.DeleteOne(ctx, bson.M{"_id": at.ID})
			if err != nil {
				return err
			}
			can0, err := getImgs(imgsC, "cancel_0")
			if err != nil {
				return err
			}
			can1, err := getImgs(imgsC, "cancel_1")
			if err != nil {
				return err
			}
			if at.From == int64(update.Message.From.ID) {
				_, err = bot.ReplyWithPhoto(update.Message.MessageID, randImg(can0), "Вы отменили атаку", update.Message.Chat.ID)
				if err != nil {
					return err
				}
				_, err = bot.SendPhoto(
					randImg(can1),
					fmt.Sprintf(
						"Вомбат %s решил вернуть вомбата домой. Вы свободны от атак",
						womb.Name,
					), at.To,
				)
				return err
			}
			_, err = bot.ReplyWithPhoto(update.Message.MessageID, randImg(can0), "Вы отклонили атаку", update.Message.Chat.ID)
			if err != nil {
				return err
			}
			_, err = bot.SendPhoto(randImg(can1),
				fmt.Sprintf(
					"Вомбат %s вежливо отказал вам в войне. Вам пришлось забрать вомбата обратно. Вы свободны от атак",
					womb.Name,
				), at.From,
			)
			return err
		},
	},
	{
		Name: "acccept",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "принять"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if len(args) > 2 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Атака принять: слишком много аргументов", update.Message.Chat.ID)
				return err
			} else if !isInUsers {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Но у вас вомбата нет...", update.Message.Chat.ID)
				return err
			}
			var at Attack
			if is, isFrom := isInAttacks(update.Message.From.ID, attacks); isFrom {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Ну ты чо... атаку принимает тот, кого атакуют...", update.Message.Chat.ID)
				return err
			} else if is {
				a, err := getAttackByWomb(update.Message.From.ID, false, attacks)
				if err != nil {
					return err
				}
				at = a
			} else {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Вам нечего принимать...", update.Message.Chat.ID)
				return err
			}
			rCount, err := users.CountDocuments(ctx, bson.M{"_id": at.From})
			if err != nil {
				return err
			} else if rCount < 1 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID,
					"Ну ты чаво... Соперника не существует! Как вообще мы такое допустили?! (ответь на это командой /admin)",
					update.Message.Chat.ID,
				)
				return err
			}
			var tWomb User
			err = users.FindOne(ctx, bson.M{"_id": at.From}).Decode(&tWomb)
			if err != nil {
				return err
			}
			atimgs, err := getImgs(imgsC, "attacks")
			if err != nil {
				return err
			}
			im := randImg(atimgs)
			ph1, err := bot.ReplyWithPhoto(update.Message.MessageID, im, "", update.Message.Chat.ID)
			if err != nil {
				return err
			}
			ph2, err := bot.SendPhoto(im, "", tWomb.ID)
			if err != nil {
				return err
			}
			war1, err := bot.ReplyWithMessage(ph1, "Да начнётся вомбой!", update.Message.Chat.ID)
			if err != nil {
				return err
			}
			war2, err := bot.ReplyWithMessage(ph2, fmt.Sprintf(
				"АААА ВАЙНААААА!!!\n Вомбат %s всё же принял ваше предложение",
				womb.Name), tWomb.ID,
			)
			if err != nil {
				return err
			}
			time.Sleep(5 * time.Second)
			h1, h2 := int(womb.Health), int(tWomb.Health)
			for _, round := range []int{1, 2, 3} {
				f1 := uint32(2 + rand.Intn(int(womb.Force-1)))
				f2 := uint32(2 + rand.Intn(int(tWomb.Force-1)))
				err = bot.EditMessage(war1, fmt.Sprintf(
					"РАУНД %d\n\nВаш вомбат:\n - здоровье: %d\n -Ваш удар: %d\n\n%s:\n - здоровье: %d",
					round, h1, f1, tWomb.Name, h2), update.Message.Chat.ID,
				)
				if err != nil {
					return err
				}
				err = bot.EditMessage(war2, fmt.Sprintf(
					"РАУНД %d\n\nВаш вомбат:\n - здоровье: %d\n - Ваш удар: %d\n\n%s:\n - здоровье: %d",
					round, h2, f2, womb.Name, h1), tWomb.ID,
				)
				if err != nil {
					return err
				}
				time.Sleep(3 * time.Second)
				h1 -= int(f2)
				h2 -= int(f1)
				err = bot.EditMessage(war1, fmt.Sprintf(
					"РАУНД %d\n\nВаш вомбат:\n - здоровье: %d\n - Ваш удар: %d\n\n%s:\n - здоровье: %d\n - 💔 удар: %d",
					round, h1, f1, tWomb.Name, h2, f2), update.Message.Chat.ID,
				)
				if err != nil {
					return err
				}
				err = bot.EditMessage(war2, fmt.Sprintf(
					"РАУНД %d\n\nВаш вомбат:\n - здоровье: %d\n - Ваш удар: %d\n\n%s:\n - здоровье: %d\n - 💔 удар: %d",
					round, h2, f2, womb.Name, h1, f1), tWomb.ID,
				)
				if err != nil {
					return err
				}
				time.Sleep(5 * time.Second)
				if int(h2)-int(f1) <= 5 && int(h1)-int(f2) <= 5 {
					err = bot.EditMessage(war1,
						"Вы оба сдохли!!!)\nВаши характеристики не поменялись, но зато да.",
						update.Message.Chat.ID,
					)
					if err != nil {
						return err
					}
					err = bot.EditMessage(war2,
						"Вы оба сдохли!!!)\nВаши характеристики не поменялись, но зато да.",
						tWomb.ID,
					)
					if err != nil {
						return err
					}
					time.Sleep(5 * time.Second)
					break
				} else if int(h2)-int(f1) <= 5 {
					err = bot.EditMessage(war1, fmt.Sprintf(
						"В раунде %d благодаря своей силе победил вомбат...",
						round), update.Message.Chat.ID,
					)
					if err != nil {
						return err
					}
					err = bot.EditMessage(war2, fmt.Sprintf(
						"В раунде %d благодаря лишению у другого здоровья победил вомбат...",
						round), tWomb.ID,
					)
					if err != nil {
						return err
					}
					time.Sleep(3 * time.Second)
					h1c := int(womb.Health) / ((5 + rand.Intn(5)) / (rand.Intn(1) + 1))
					f1c := int(womb.Force) / ((5 + rand.Intn(5)) / (rand.Intn(1) + 1))
					mc := int((rand.Intn(int(womb.Health)) + 1) / 2)
					womb.Health += uint32(h1c)
					womb.Force += uint32(f1c)
					womb.Money += uint32(mc)
					womb.XP += 10
					err = bot.EditMessage(war1, fmt.Sprintf(
						"Победил вомбат %s!!!\nВы получили 10 XP, %d силы, %d здоровья и %d шишей, теперь их у Вас %d, %d, %d и %d соответственно",
						womb.Name, h1c, f1c, mc, womb.XP, womb.Health, womb.Force, womb.Money), update.Message.Chat.ID,
					)
					if err != nil {
						return err
					}
					tWomb.Health = 5
					tWomb.Money = 50
					err = bot.EditMessage(war2, fmt.Sprintf(
						"Победил вомбат %s!!!\nВаше здоровье обнулилось, а ещё у вас теперь только 50 шишей при себе :(",
						womb.Name), tWomb.ID,
					)
					if err != nil {
						return err
					}
					break
				} else if int(h1)-int(f2) <= 5 {
					err = bot.EditMessage(war1, fmt.Sprintf(
						"В раунде %d благодаря своей силе победил вомбат...",
						round), update.Message.Chat.ID,
					)
					if err != nil {
						return err
					}
					err = bot.EditMessage(war2, fmt.Sprintf(
						"В раунде %d благодаря лишению у другого здоровья победил вомбат...",
						round), tWomb.ID,
					)
					if err != nil {
						return err
					}
					time.Sleep(3 * time.Second)
					h2c := int(tWomb.Health) / ((5 + rand.Intn(5)) / (rand.Intn(1) + 1))
					f2c := int(tWomb.Force) / ((5 + rand.Intn(5)) / (rand.Intn(1) + 1))
					mc := int((rand.Intn(int(tWomb.Health)) + 1) / 2)
					tWomb.Health += uint32(h2c)
					tWomb.Force += uint32(f2c)
					tWomb.Money += uint32(mc)
					tWomb.XP += 10
					err = bot.EditMessage(war2,
						fmt.Sprintf(
							"Победил вомбат %s!!!\nВы получили 10 XP, %d силы, %d здоровья и %d шишей, теперь их у Вас %d, %d, %d и %d соответственно",
							tWomb.Name, h2c, f2c, mc, tWomb.XP, tWomb.Health, tWomb.Force, tWomb.Money,
						), tWomb.ID,
					)
					if err != nil {
						return err
					}
					womb.Health = 5
					womb.Money = 50
					err = bot.EditMessage(war1,
						fmt.Sprintf(
							"Победил вомбат %s!!!\nВаше здоровье сбросилось до 5, а ещё у вас теперь только 50 шишей при себе :(",
							tWomb.Name,
						),
						update.Message.Chat.ID,
					)
					if err != nil {
						return err
					}
					break
				} else if round == 3 {
					if h1 < h2 {
						h2c := int(tWomb.Health) / ((5 + rand.Intn(5)) / (rand.Intn(1) + 1))
						f2c := int(tWomb.Force) / ((5 + rand.Intn(5)) / (rand.Intn(1) + 1))
						mc := int((rand.Intn(int(tWomb.Health)) + 1) / 2)
						tWomb.Health += uint32(h2c)
						tWomb.Force += uint32(f2c)
						tWomb.Money += uint32(mc)
						tWomb.XP += 10
						err = bot.EditMessage(war2,
							fmt.Sprintf(
								"И победил вомбат %s на раунде %d!!!\nВы получили 10 XP, %d силы, %d здоровья и %d шишей, теперь их у Вас %d, %d, %d и %d соответственно",
								tWomb.Name, round, h2c, f2c, mc, tWomb.XP, tWomb.Health, tWomb.Force, tWomb.Money,
							),
							tWomb.ID,
						)
						if err != nil {
							return err
						}
						womb.Health = uint32(h1)
						womb.Money = 50
						err = bot.EditMessage(war1,
							fmt.Sprintf(
								"И победил вомбат %s на раунде %d!\n К сожалению, теперь у вас только %d здоровья и 50 шишей при себе :(",
								tWomb.Name, round, womb.Health,
							),
							update.Message.Chat.ID,
						)
						if err != nil {
							return err
						}
					} else {
						h1c := int(womb.Health) / ((5 + rand.Intn(5)) / (rand.Intn(1) + 1))
						f1c := int(womb.Force) / ((5 + rand.Intn(5)) / (rand.Intn(1) + 1))
						mc := int((rand.Intn(int(womb.Health)) + 1) / 2)
						womb.Health += uint32(h1c)
						womb.Force += uint32(f1c)
						womb.Money += uint32(mc)
						womb.XP += 10
						err = bot.EditMessage(war1,
							fmt.Sprintf(
								"Победил вомбат %s!!!\nВы получили 10 XP, %d силы, %d здоровья и %d шишей, теперь их у Вас %d, %d, %d и %d соответственно",
								womb.Name, h1c, f1c, mc, womb.XP, womb.Health, womb.Force, womb.Money,
							),
							update.Message.Chat.ID,
						)
						if err != nil {
							return err
						}
						tWomb.Health = 5
						tWomb.Money = 50
						err = bot.EditMessage(war2,
							fmt.Sprintf(
								"Победил вомбат %s!!!\nВаше здоровье обнулилось, а ещё у вас теперь только 50 шишей при себе :(",
								womb.Name,
							),
							tWomb.ID,
						)
						if err != nil {
							return err
						}
					}
				}
			}
			err = docUpd(womb, wombFilter(womb), users)
			if err != nil {
				return err
			}
			err = docUpd(tWomb, bson.M{"_id": tWomb.ID}, users)
			if err != nil {
				return err
			}
			_, err = attacks.DeleteOne(ctx, bson.M{"_id": at.ID})
			return err
		},
	},
}

var bankCommands = []command{
	{
		Name: "wombank",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "вомбанк"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			_, err := bot.ReplyWithMessage(update.Message.MessageID, strings.Repeat("вомбанк ", 42), update.Message.Chat.ID)
			return err
		},
	},
	{
		Name: "new",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "начать"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			isBanked, err := getIsBanked(update.Message.From.ID)
			if err != nil {
				return err
			}
			if len(args) != 2 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Вомбанк начать: слишком много аргументов", update.Message.Chat.ID)
				return err
			} else if isBanked {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Ты уже зарегестрирован в вомбанке...", update.Message.Chat.ID)
				return err
			} else if !isInUsers {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Вомбанк вомбатам! У тебя нет вомбата", update.Message.Chat.ID)
				return err
			}
			b := Banked{
				ID:    update.Message.From.ID,
				Money: 15,
			}
			_, err = bank.InsertOne(ctx, b)
			if err != nil {
				return err
			}
			_, err = bot.ReplyWithMessage(
				update.Message.MessageID,
				"Вы были зарегестрированы в вомбанке! Вам на вомбосчёт добавили бесплатные 15 шишей",
				update.Message.Chat.ID,
			)
			return err
		},
	},
	{
		Name: "put",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "положить"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			isBanked, err := getIsBanked(update.Message.From.ID)
			if err != nil {
				return err
			}
			if !isInUsers {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "У тебя нет вомбата...", update.Message.Chat.ID)
				return err
			} else if len(args) != 3 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Вомбанк положить: недостаточно аргументов", update.Message.Chat.ID)
				return err
			}
			var num uint64
			if num, err = strconv.ParseUint(args[2], 10, 64); err != nil {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Вомбанк положить: требуется целое неотрицательное число шишей до 2^64", update.Message.Chat.ID)
				return err
			}
			if womb.Money < uint32(num)+1 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Вомбанк положить: недостаточно шишей при себе для операции", update.Message.Chat.ID)
				return err
			} else if !isBanked {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Вомбанк положить: у вас нет ячейки в банке! Заведите её через `вомбанк начать`",
					update.Message.Chat.ID,
				)
				return err
			} else if num == 0 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Ну и зачем?)", update.Message.Chat.ID)
				return err
			}
			var b Banked
			err = bank.FindOne(ctx, wombFilter(womb)).Decode(&b)
			if err != nil {
				return err
			}
			womb.Money -= uint32(num)
			b.Money += uint32(num)
			err = docUpd(womb, wombFilter(womb), users)
			if err != nil {
				return err
			}
			err = docUpd(b, wombFilter(womb), bank)
			if err != nil {
				return err
			}
			_, err = bot.ReplyWithMessage(
				update.Message.MessageID,
				fmt.Sprintf(
					"Ваш вомбосчёт пополнен на %d ш! Вомбосчёт: %d ш; При себе: %d ш",
					num, b.Money, womb.Money,
				),
				update.Message.Chat.ID,
			)
			return err
		},
	},
	{
		Name: "take",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "снять"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			isBanked, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if !isInUsers {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "У тебя нет вомбата...", update.Message.Chat.ID)
				return err
			} else if !isBanked {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "у тебя нет ячейки в вомбанке", update.Message.Chat.ID)
				return err
			} else if len(args) != 3 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Вомбанк снять: недостаточно аргументов", update.Message.Chat.ID)
				return err
			}
			var b Banked
			err = bank.FindOne(ctx, wombFilter(womb)).Decode(&b)
			if err != nil {
				return err
			}
			var num uint64
			if num, err = strconv.ParseUint(args[2], 10, 64); err == nil {
				if num == 0 {
					_, err = bot.ReplyWithMessage(update.Message.MessageID, "Ну и зачем?", update.Message.Chat.ID)
					return err
				}
			} else if args[2] == "всё" || args[2] == "все" {
				if b.Money == 0 {
					_, err = bot.ReplyWithMessage(update.Message.MessageID, "У вас на счету 0 шишей. Зачем?", update.Message.Chat.ID)
					return err
				}
				num = uint64(b.Money)
			} else {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Вомбанк снять: требуется целое неотрицательное число шишей до 2^64", update.Message.Chat.ID)
				return err
			}
			if b.Money < uint32(num) {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Вомбанк снять: недостаточно шишей на вомбосчету для операции", update.Message.Chat.ID)
				return err
			}
			b.Money -= uint32(num)
			womb.Money += uint32(num)
			err = docUpd(b, wombFilter(womb), bank)
			if err != nil {
				return err
			}
			err = docUpd(womb, wombFilter(womb), users)
			if err != nil {
				return err
			}
			_, err = bot.ReplyWithMessage(
				update.Message.MessageID,
				fmt.Sprintf(
					"Вы сняли %d ш! Вомбосчёт: %d ш; При себе: %d ш",
					num, b.Money, womb.Money,
				),
				update.Message.Chat.ID,
			)
			return err
		},
	},
	{
		Name: "status",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "статус"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			var (
				fil   bson.M
				tWomb User
				err   error
			)
			switch len(args) {
			case 2:
				isInUsers, err := getIsInUsers(update.Message.From.ID)
				if err != nil {
					return err
				}
				isBanked, err := getIsBanked(update.Message.From.ID)
				if err != nil {
					return err
				}
				if !isInUsers {
					_, err = bot.ReplyWithMessage(update.Message.MessageID, "Вомбанк вомбатам! У тебя нет вомбата", update.Message.Chat.ID)
					return err
				} else if !isBanked {
					_, err = bot.ReplyWithMessage(update.Message.MessageID, "Вы не можете посмотреть вомбосчёт, которого нет", update.Message.Chat.ID)
					return err
				}
				fil = bson.M{"_id": update.Message.From.ID}
				tWomb = womb
			case 3:
				name := args[2]
				if !isValidName(name) {
					_, err = bot.ReplyWithMessage(update.Message.MessageID, "Нелегальное имя", update.Message.Chat.ID)
					return err
				} else if rCount, err := users.CountDocuments(
					ctx, bson.M{"name": cins(name)}); err != nil {
					return err
				} else if rCount == 0 {
					_, err = bot.ReplyWithMessage(update.Message.MessageID, fmt.Sprintf("Ошибка: вомбата с именем %s не найдено", name), update.Message.Chat.ID)
					return err
				}
				err = users.FindOne(ctx, bson.M{"name": cins(name)}).Decode(&tWomb)
				if err != nil {
					return err
				}
				fil = bson.M{"_id": tWomb.ID}
				bCount, err := bank.CountDocuments(ctx, fil)
				if err != nil {
					return err
				}
				if bCount == 0 {
					_, err = bot.ReplyWithMessage(
						update.Message.MessageID,
						"Ошибка: вомбат с таким именем не зарегестрирован в вомбанке",
						update.Message.Chat.ID,
					)
					return err
				}
			default:
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Вомбанк статус: слишком много аргументов", update.Message.Chat.ID)
				return err
			}
			var b Banked
			err = bank.FindOne(ctx, fil).Decode(&b)
			if err != nil {
				return err
			}
			_, err = bot.ReplyWithMessage(
				update.Message.MessageID,
				fmt.Sprintf(
					"Вомбанк вомбата %s:\nНа счету: %d\nПри себе: %d",
					tWomb.Name, b.Money, tWomb.Money,
				),
				update.Message.Chat.ID,
			)
			return err
		},
	},
}

var clanCommands = []command{
	{
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "клан"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			_, err := bot.ReplyWithMessage(update.Message.MessageID, strings.Repeat("атака ", 42), update.Message.Chat.ID)
			return err
		},
	},
	{
		Name: "new",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "создать"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if !isInUsers {
				_, err := bot.ReplyWithMessage(
					update.Message.MessageID,
					"Кланы - приватная территория вомбатов. У тебя вомбата нет",
					update.Message.Chat.ID,
				)
				return err
			} else if len(args) < 4 {
				_, err := bot.ReplyWithMessage(
					update.Message.MessageID,
					"Клан создать: недостаточно аргументов. Синтаксис: клан создать "+
						"[тег (3-5 латинские буквы)] [имя (можно пробелы)]",
					update.Message.Chat.ID,
				)
				return err
			} else if womb.Money < 25000 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					fmt.Sprintf(
						"Ошибка: недостаточно шишей. Требуется 25'000 шишей при себе для создания клана (У вас их при себе %d)",
						womb.Money,
					),
					update.Message.Chat.ID,
				)
				return err
			} else if l := len([]rune(args[2])); !(l >= 3 && l <= 5) {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Слишком длинный тэг!", update.Message.Chat.ID)
				return err
			} else if !isValidTag(args[2]) {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Нелегальный тэг(", update.Message.Chat.ID)
				return err
			} else if name := strings.Join(args[3:], " "); len([]rune(name)) > 64 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Слишком длинное имя! Оно должно быть максимум 64 символов",
					update.Message.Chat.ID,
				)
				return err
			} else if len([]rune(name)) < 2 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Слишком короткое имя! Оно должно быть минимум 3 символа",
					update.Message.Chat.ID,
				)
				return err
			}
			tag, name := strings.ToLower(args[2]), strings.Join(args[3:], " ")
			if rCount, err := clans.CountDocuments(ctx,
				bson.M{"_id": cins(tag)}); err != nil {
				return err
			} else if rCount != 0 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					fmt.Sprintf(
						"Ошибка: клан с тегом `%s` уже существует",
						tag,
					),
					update.Message.Chat.ID,
				)
				return err
			}
			if rCount, err := clans.CountDocuments(ctx,
				bson.M{"members": update.Message.From.ID}); err != nil {
				return err
			} else if rCount != 0 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Ошибка: вы уже состоите в клане. Напишите `клан выйти`, чтобы выйти из него",
					update.Message.Chat.ID,
				)
				return err
			}
			womb.Money -= 25000
			err = docUpd(womb, wombFilter(womb), users)
			if err != nil {
				return err
			}
			nclan := Clan{
				Tag:     strings.ToUpper(tag),
				Name:    name,
				Money:   100,
				Leader:  update.Message.From.ID,
				Banker:  update.Message.From.ID,
				Members: []int64{update.Message.From.ID},
				Banned:  []int64{},
				GroupID: update.Message.Chat.ID,
				Settings: ClanSettings{
					AviableToJoin: true,
				},
			}
			_, err = clans.InsertOne(ctx, &nclan)
			if err != nil {
				return err
			}
			if hasTitle(5, womb.Titles) {
				newTitles := []uint16{}
				for _, id := range womb.Titles {
					if id == 5 {
						continue
					}
					newTitles = append(newTitles, id)
				}
				womb.Titles = newTitles
				err = docUpd(womb, wombFilter(womb), users)
				if err != nil {
					return err
				}
			}
			_, err = bot.ReplyWithMessage(
				update.Message.MessageID,
				fmt.Sprintf(
					"Клан `%s` успешно создан и привязан к этой группе! У вас взяли 25'000 шишей",
					name,
				),
				update.Message.Chat.ID,
			)
			return err
		},
	},
	{
		Name: "join",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "вступить"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if !isInUsers {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Кланы - приватная территория вомбатов. Вомбата у тебя нет.",
					update.Message.Chat.ID,
				)
				return err
			} else if len(args) != 3 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Клан вступить: слишком мало или много аргументов! Синтаксис: клан вступить [тэг клана]",
					update.Message.Chat.ID,
				)
				return err
			} else if womb.Money < 1000 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Клан вступить: недостаточно шишей (надо минимум 1000 ш)",
					update.Message.Chat.ID,
				)
				return err
			} else if rCount, err := clans.CountDocuments(ctx,
				bson.M{"members": update.Message.MessageID}); err != nil {
				return err
			} else if rCount != 0 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Ошибка: вы уже состоите в клане. Напишите `клан выйти`, чтобы выйти из него",
					update.Message.Chat.ID,
				)
				return err
			} else if l := len([]rune(args[2])); !(l >= 3 && l <= 5) {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Слишком длинный или короткий тег :)", update.Message.Chat.ID)
				return err
			} else if !isValidTag(args[2]) {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Тег нелгальный(", update.Message.Chat.ID)
				return err
			} else if rCount, err := clans.CountDocuments(ctx,
				bson.M{"_id": strings.ToUpper(args[2])}); err != nil {
				return err
			} else if rCount == 0 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					fmt.Sprintf(
						"Ошибка: клана с тегом `%s` не существует",
						args[2],
					),
					update.Message.Chat.ID,
				)
				return err
			}
			var jClan Clan
			err = clans.FindOne(ctx, bson.M{"_id": strings.ToUpper(args[2])}).Decode(&jClan)
			if err != nil {
				return err
			}
			if len(jClan.Members) >= 7 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Ошибка: в клане слишком много игроков :(", update.Message.Chat.ID)
				return err
			} else if !(jClan.Settings.AviableToJoin) {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "К сожалению, клан закрыт для вступления", update.Message.Chat.ID)
				return err
			} else if update.Message.Chat.ID != jClan.GroupID {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Для вступления в клан Вы должны быть в зарегестрированном чате клана",
					update.Message.Chat.ID,
				)
				return err
			}
			for _, id := range jClan.Banned {
				if id == womb.ID {
					_, err = bot.ReplyWithMessage(update.Message.MessageID, "Вы забанены!!1\n в этом клане(", update.Message.Chat.ID)
					return err
				}
			}
			womb.Money -= 1000
			err = docUpd(womb, wombFilter(womb), users)
			if err != nil {
				return err
			}
			jClan.Members = append(jClan.Members, update.Message.From.ID)
			err = docUpd(jClan, bson.M{"_id": strings.ToUpper(args[2])}, clans)
			if err != nil {
				return err
			}
			_, err = bot.ReplyWithMessage(
				update.Message.MessageID,
				"Отлично, вы присоединились! У вас взяли 1000 шишей",
				update.Message.Chat.ID,
			)
			if err != nil {
				return err
			}
			_, err = bot.SendMessage(
				fmt.Sprintf(
					"В ваш клан вступил вомбат `%s`",
					womb.Name,
				),
				jClan.Leader,
			)
			return err
		},
	},
	{
		Name: "set_user",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "назначить"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if len(args) == 2 {
				_, err := bot.ReplyWithMessage(update.Message.MessageID, "конечно", update.Message.Chat.ID)
				return err
			}
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			switch args[2] {
			case "назначить":
				_, err = bot.ReplyWithMessage(update.Message.MessageID, strings.Repeat("назначить", 42), update.Message.Chat.ID)
				return err
			case "лидера":
				fallthrough
			case "лидером":
				fallthrough
			case "лидер":
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Используйте \"клан передать [имя]\" вместо данной команды", update.Message.Chat.ID)
				return err
			case "казначея":
				fallthrough
			case "казначеем":
				fallthrough
			case "казначей":
				if len(args) != 4 {
					_, err = bot.ReplyWithMessage(update.Message.MessageID, "Слишком много или мало аргументов", update.Message.Chat.ID)
					return err
				} else if !isInUsers {
					_, err = bot.ReplyWithMessage(
						update.Message.MessageID,
						"Кланы — приватная территория вомбатов. У тебя вомбата нет.",
						update.Message.Chat.ID,
					)
					return err
				}
				if c, err := clans.CountDocuments(ctx, bson.M{"leader": update.Message.From.ID}); err != nil {
					return err
				} else if c == 0 {
					_, err = bot.ReplyWithMessage(
						update.Message.MessageID,
						"Вы не состоите ни в одном клане либо не являетесь лидером клана",
						update.Message.Chat.ID,
					)
					return err
				}
				var sClan Clan
				if err := clans.FindOne(ctx, bson.M{"leader": update.Message.From.ID}).Decode(&sClan); err != nil {
					return err
				}
				lbid := sClan.Banker
				name := args[3]
				if c, err := users.CountDocuments(ctx, bson.M{"name": cins(name)}); err != nil {
					return err
				} else if c == 0 {
					_, err = bot.ReplyWithMessage(
						update.Message.MessageID,
						"Вомбата с таким ником не найдено",
						update.Message.Chat.ID,
					)
					return err
				}
				var (
					nb User
				)
				if err := users.FindOne(ctx, bson.M{"name": cins(name)}).Decode(&nb); err != nil {
					return err
				}
				var is bool
				for _, id := range sClan.Members {
					if id == nb.ID {
						is = true
						break
					}
				}
				if !is {
					_, err = bot.ReplyWithMessage(update.Message.MessageID, "Данный вобат не состоит в Вашем клане", update.Message.Chat.ID)
					return err
				}
				sClan.Banker = nb.ID
				if err := docUpd(sClan, bson.M{"_id": sClan.Tag}, clans); err != nil {
					return err
				}
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Казначей успешно изменён! Теперь это "+nb.Name,
					update.Message.Chat.ID,
				)
				if err != nil {
					return err
				}
				if nb.ID != update.Message.From.ID {
					_, err = bot.SendMessage("Вы стали казначеем в клане `"+sClan.Name+"` ["+sClan.Tag+"]", nb.ID)
					if err != nil {
						return err
					}
				}
				if lbid != update.Message.From.ID && lbid != 0 {
					_, err = bot.SendMessage("Вы казначей... теперь бывший. (в клане `"+sClan.Name+"` ["+sClan.Tag+"])", lbid)
					return err
				}
				return nil
			default:
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Не знаю такой роли в клане(", update.Message.Chat.ID)
				return err
			}
		},
	},
	{
		Name: "transfer",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "передать"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if len(args) != 3 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Ошибка: слишком много или мало аргументов. Синтаксис: клан передать [ник]",
					update.Message.Chat.ID,
				)
				return err
			} else if !isInUsers {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Кланы — приватная территория вомбатов. У тебя вомбата (хлюп) нет",
					update.Message.Chat.ID,
				)
				return err
			} else if rCount, err := clans.CountDocuments(ctx,
				bson.M{"leader": update.Message.From.ID}); err != nil {
				return err
			} else if rCount == 0 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Ошибка: вы не лидер ни в одном клане!!!11!!!",
					update.Message.Chat.ID,
				)
				return err
			} else if rCount, err := users.CountDocuments(ctx,
				bson.M{"name": cins(args[2])}); err != nil {
				return err
			} else if rCount == 0 {
				bot.ReplyWithMessage(
					update.Message.MessageID,
					"Ошибка: пользователя с таким ником не найдено",
					update.Message.Chat.ID,
				)
				return err
			}
			var newLead User
			err = users.FindOne(ctx, bson.M{"name": cins(args[2])}).Decode(&newLead)
			if err != nil {
				return err
			}
			if strings.ToLower(args[2]) == strings.ToLower(womb.Name) {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Но ты и так лидер...", update.Message.Chat.ID)
				return err
			} else if rCount, err := clans.CountDocuments(ctx,
				bson.M{"members": newLead.ID}); err != nil {
				return err
			} else if rCount == 0 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					fmt.Sprintf("Ошибка: вомбат `%s` не состоит ни в одном клане", newLead.Name),
					update.Message.Chat.ID,
				)
				return err
			}
			var uClan Clan
			err = clans.FindOne(ctx, bson.M{"leader": update.Message.From.ID}).Decode(&uClan)
			if err != nil {
				return err
			}
			var isIn bool = false
			for _, id := range uClan.Members {
				if id == newLead.ID {
					isIn = true
					break
				}
			}
			if !isIn {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					fmt.Sprintf("Ошибка: вы и %s состоите в разных кланах", newLead.Name),
					update.Message.Chat.ID,
				)
				return err
			}
			uClan.Leader = newLead.ID
			err = docUpd(uClan, bson.M{"_id": uClan.Tag}, clans)
			if err != nil {
				return err
			}
			if hasTitle(5, womb.Titles) {
				newTitles := []uint16{}
				for _, id := range womb.Titles {
					if id == 5 {
						continue
					}
					newTitles = append(newTitles, id)
				}
				womb.Titles = newTitles
				if err != nil {
					return err
				}
			}
			if !hasTitle(5, newLead.Titles) {
				newLead.Titles = append(newLead.Titles, 5)
				err = docUpd(newLead, bson.M{"_id": newLead.ID}, users)
				if err != nil {
					return err
				}
			}
			_, err = bot.ReplyWithMessage(
				update.Message.MessageID,
				fmt.Sprintf(
					"Отлично! Вомбат `%s` теперь главный в клане `%s`",
					newLead.Name, uClan.Tag,
				),
				update.Message.Chat.ID,
			)
			if err != nil {
				return err
			}
			_, err = bot.SendMessage("Вам передали права на клан!", newLead.ID)
			return err
		},
	},
	{
		Name: "quit",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "выйти"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if !isInUsers {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Кланы — приватная территория вомбатов. У тебя вомбата нет (мне уже надоело это писать в каждом сообщении, заведи уже вомбата нафек)",
					update.Message.Chat.ID,
				)
				return err
			} else if len(args) != 2 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Ошибка: слишком много или мало аргументов. Синтаксис: клан выйти",
					update.Message.Chat.ID,
				)
				return err
			} else if rCount, err := clans.CountDocuments(ctx,
				bson.M{"members": update.Message.From.ID}); err != nil {
				return err
			} else if rCount == 0 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Клан выйти: вы не состоите ни в одном клане", update.Message.Chat.ID)
				return err
			}
			var uClan Clan
			err = clans.FindOne(ctx, bson.M{"members": update.Message.From.ID}).Decode(&uClan)
			if err != nil {
				return err
			}
			if len(uClan.Members) == 1 {
				_, err = clans.DeleteOne(ctx, bson.M{"_id": uClan.Tag})
				if err != nil {
					return err
				}
				if uClan.Leader == update.Message.From.ID {
					if hasTitle(5, womb.Titles) {
						newTitles := []uint16{}
						for _, id := range womb.Titles {
							if id == 5 {
								continue
							}
							newTitles = append(newTitles, id)
						}
						womb.Titles = newTitles
						err = docUpd(womb, wombFilter(womb), users)
						if err != nil {
							return err
						}
					}
				}
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Так как вы были одни в клане, то клан удалён",
					update.Message.Chat.ID,
				)
				return err
			} else if uClan.Leader == update.Message.From.ID {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Клан выйти: вы лидер. Передайте кому-либо ваши права",
					update.Message.Chat.ID,
				)
				return err
			}
			newMembers := []int64{}
			for _, id := range uClan.Members {
				if id == update.Message.From.ID {
					continue
				}
				newMembers = append(newMembers, id)
			}
			var (
				rep    string = "Вы вышли из клана. Вы свободны!"
				msgtol string = "Вомбат `" + womb.Name + "` вышел из клана."
			)
			uClan.Members = newMembers
			if uClan.Banker == update.Message.From.ID && uClan.Leader != uClan.Banker {
				uClan.Banker = uClan.Leader
				rep += "\nБанкиром вместо вас стал лидер клана."
				msgtol += "\nТак как этот вомбат был банкиром, Вы стали банкиром клана."
			}
			err = docUpd(uClan, bson.M{"_id": uClan.Tag}, clans)
			if err != nil {
				return err
			}
			_, err = bot.ReplyWithMessage(
				update.Message.MessageID,
				rep,
				update.Message.Chat.ID,
			)
			if err != nil {
				return err
			}
			_, err = bot.SendMessage(msgtol, uClan.Leader)
			if err != nil {
				return err
			}
			if update.Message.Chat.ID != uClan.GroupID {
				_, err = bot.SendMessage("Вомбат "+womb.Name+" вышел из клана.", uClan.GroupID)
				if err != nil {
					return err
				}
			}
			return nil
		},
	},
	{
		Name: "status",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "статус"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if len(args) > 3 {
				_, err := bot.ReplyWithMessage(
					update.Message.MessageID,
					"Клан статус: слишком много аргументов! Синтаксис: клан статус ([тег])",
					update.Message.Chat.ID,
				)
				return err
			}
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			var sClan Clan
			if len(args) == 2 {
				if !isInUsers {
					_, err = bot.ReplyWithMessage(
						update.Message.MessageID,
						"Вы не имеете вомбата. Соответственно, вы не состоите в ни в одном вомбоклане",
						update.Message.Chat.ID,
					)
					return err
				} else if rCount, err := clans.CountDocuments(ctx,
					bson.M{"members": update.Message.From.ID}); err != nil {
					return err
				} else if rCount == 0 {
					_, err = bot.ReplyWithMessage(
						update.Message.MessageID,
						"Клан статус: вы не состоите ни в одном клане",
						update.Message.Chat.ID,
					)
					return err
				}
				err = clans.FindOne(ctx, bson.M{"members": update.Message.From.ID}).Decode(&sClan)
				if err != nil {
					return err
				}
			} else {
				if l := len([]rune(args[2])); !(l >= 3 && l <= 5) {
					_, err = bot.ReplyWithMessage(
						update.Message.MessageID,
						"Ошибка: слишком длинный или короткий тег",
						update.Message.Chat.ID,
					)
					return err
				} else if !isValidTag(args[2]) {
					_, err = bot.ReplyWithMessage(
						update.Message.MessageID,
						"Ошибка: тег нелегален",
						update.Message.Chat.ID,
					)
					return err
				} else if rCount, err := clans.CountDocuments(ctx,
					bson.M{"_id": cins(args[2])}); err != nil {
					return err
				} else if rCount == 0 {
					_, err = bot.ReplyWithMessage(
						update.Message.MessageID,
						fmt.Sprintf(
							"Ошибка: клана с тегом `%s` не существует",
							args[2],
						),
						update.Message.Chat.ID,
					)
					return err
				}
				err = clans.FindOne(ctx, bson.M{"_id": cins(args[2])}).Decode(&sClan)
				if err != nil {
					return err
				}
			}
			var (
				msg string = fmt.Sprintf(
					"Клан `%s` [%s]\n 💰 Казна: %d шишей\n 🐽 Участники:\n",
					sClan.Name, sClan.Tag, sClan.Money,
				)
				midHealth, midForce uint32
				lost                uint8 = 0
			)
			var tWomb User
			for i, id := range sClan.SortedMembers() { // append для порядка
				if rCount, err := users.CountDocuments(ctx,
					bson.M{"_id": id}); err != nil {
					return err
				} else if rCount == 0 {
					msg += " - Вомбат не найден :("
					lost++
					continue
				} else {
					err = users.FindOne(ctx, bson.M{"_id": id}).Decode(&tWomb)
					if err != nil {
						return err
					}
					msg += fmt.Sprintf("        %d. [%s](tg://user?id=%d)", i+1, tWomb.Name, tWomb.ID)
					if id == sClan.Leader {
						msg += " | Лидер"
					} else if sClan.Banker == id {
						msg += " | Казначей"
					}
					midHealth += tWomb.Health
					midForce += tWomb.Force
				}
				msg += "\n"
			}
			if uint32(len(sClan.Members)-int(lost)) != 0 {
				midHealth /= uint32(len(sClan.Members) - int(lost))
				midForce /= uint32(len(sClan.Members) - int(lost))
			} else {
				midHealth, midForce = 0, 0
			}
			msg += fmt.Sprintf(
				" ❤ Среднее здоровье: %d\n ⚡ Средняя мощь: %d\n 👁 XP: %d",
				midHealth, midForce, sClan.XP,
			)
			_, err = bot.ReplyWithMessage(update.Message.MessageID, msg, update.Message.Chat.ID, MarkdownParseModeMessage)
			return err
		},
	},
	{
		Name: "award",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "награда"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if !isInUsers {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Кланы — приватная территория вомбатов. У тебя вомбата нет.",
					update.Message.Chat.ID,
				)
				return err
			}
			if c, err := clans.CountDocuments(ctx, bson.M{"members": update.Message.From.ID}); err != nil {
				return err
			} else if c == 0 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Вы не состоите ни в одном клане", update.Message.Chat.ID)
				return err
			}
			var sClan Clan
			if err := clans.FindOne(ctx, bson.M{"members": update.Message.From.ID}).Decode(&sClan); err != nil {
				return err
			}
			if !(update.Message.From.ID == sClan.Leader || update.Message.From.ID == sClan.Banker) {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Для того, чтобы получить награду, вы должны быть казначеем или лидером",
					update.Message.Chat.ID,
				)
				return err
			}
			if e := time.Now().Sub(sClan.LastRewardTime); e < 24*time.Hour {
				left := (24 * time.Hour) - e
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					fmt.Sprintf(
						"С момента прошлого получения награды не прошло 24 часов. "+
							"Осталось %d часов %d минут",
						int64(left.Hours()), int64(left.Minutes())-int64(left.Hours())*60,
					),
					update.Message.Chat.ID,
				)
				return err
			}
			add := 500 + rand.Intn(200) - rand.Intn(200)
			sClan.Money += uint32(add)
			sClan.LastRewardTime = time.Now()
			if err := docUpd(sClan, bson.M{"_id": sClan.Tag}, clans); err != nil {
				return err
			}
			_, err = bot.ReplyWithMessage(
				update.Message.MessageID,
				fmt.Sprintf(
					"В казну клана поступило %d шишей! Теперь их %d",
					add, sClan.Money,
				),
				update.Message.Chat.ID,
			)
			return err
		},
	},
	{
		Name: "rating",
		Is: func(args []string, update tg.Update) bool {
			lowarg := strings.ToLower(args[1])
			return lowarg == "рейтинг" || lowarg == "топ"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			var (
				name  string = "xp"
				queue int8   = -1
				err   error  // because yes
			)
			if len(args) >= 3 && len(args) < 5 {
				if isInList(args[2], []string{"шиши", "деньги", "money"}) {
					name = "money"
				} else if isInList(args[2], []string{"хп", "опыт", "xp", "хрю"}) {
					name = "xp"
				} else {
					_, err = bot.ReplyWithMessage(
						update.Message.MessageID,
						"не понимаю первого аргумента(",
						update.Message.Chat.ID,
					)
					return err
				}
				if len(args) == 4 {
					if isInList(args[3], []string{"+", "плюс", "++", "увеличение"}) {
						queue = 1
					} else if isInList(args[3], []string{"-", "минус", "--", "уменьшение"}) {
						queue = -1
					} else {
						_, err = bot.ReplyWithMessage(
							update.Message.MessageID,
							"не понимаю второго аргумента, рял",
							update.Message.From.ID,
						)
						return err
					}
				}
			} else if len(args) != 2 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Слишком много аргументов", update.Message.Chat.ID)
				return err
			}
			opts := options.Find()
			opts.SetSort(bson.M{name: queue})
			opts.SetLimit(10)
			cur, err := clans.Find(ctx, bson.M{}, opts)
			if err != nil {
				return err
			}
			var rating []Clan
			for cur.Next(ctx) {
				var cl Clan
				cur.Decode(&cl)
				rating = append(rating, cl)
			}
			var msg string = "Топ-10 кланов по "
			switch name {
			case "money":
				msg += "шишам в казне "
			case "xp":
				msg += "XP "
				// no default because idk what i should put here
			}
			msg += "в порядке "
			switch queue {
			case 1:
				msg += "увеличения:"
			case -1:
				msg += "уменьшения:"
			}
			msg += "\n"
			for num, cl := range rating {
				switch name {
				case "money":
					msg += fmt.Sprintf("%d | [%s] `%s` | %d шишей в казне\n", num+1, cl.Tag, cl.Name, cl.Money)
				case "xp":
					msg += fmt.Sprintf("%d | [%s] `%s` | %d XP\n", num+1, cl.Tag, cl.Name, cl.XP)
				}
			}
			msg = strings.TrimSuffix(msg, "\n")
			_, err = bot.ReplyWithMessage(update.Message.MessageID, msg, update.Message.Chat.ID)
			return err
		},
	},
	{
		Name: "list_banned",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "забаненные"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if !isInUsers {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Кланы — приватная территория вомбатов. У тебя вомбата нет.",
					update.Message.Chat.ID,
				)
				return err
			}
			if c, err := clans.CountDocuments(ctx, bson.M{"leader": update.Message.From.ID}); err != nil {
				return err
			} else if c == 0 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Данный раздел доступен только лидерам клана; вы не являетесь лидером.",
					update.Message.From.ID,
				)
				return err
			}
			var sClan Clan
			if err := clans.FindOne(ctx, bson.M{"leader": update.Message.From.ID}).Decode(&sClan); err != nil {
				return err
			}
			if len(sClan.Banned) == 0 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Никто не в бане!", update.Message.Chat.ID)
				return err
			}
			var msg string = "⛔ Список забаненных:\n"
			for _, id := range sClan.Banned {
				var bWomb User
				if err := users.FindOne(ctx, bson.M{"_id": id}).Decode(&bWomb); err != nil {
					msg += " — [в процессе нахождения вомбата произошла ошибка]\n"
				}
				msg += " — " + bWomb.Name + "\n"
			}
			_, err = bot.ReplyWithMessage(update.Message.MessageID, msg, update.Message.Chat.ID)
			return err
		},
	},
	{
		Name: "kick",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "кик"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if len(args) == 2 {
				_, err := bot.ReplyWithMessage(update.Message.MessageID, "кого?", update.Message.Chat.ID)
				return err
			}
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if !isInUsers {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Кланы — приватная территория вомбатов. У тебя вомбата нет.",
					update.Message.Chat.ID,
				)
				return err
			}
			if c, err := clans.CountDocuments(ctx, bson.M{"leader": update.Message.From.ID}); err != nil {
				return err
			} else if c == 0 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Вы не являетесь лидером ни одного клана", update.Message.Chat.ID)
				return err
			}
			if c, err := users.CountDocuments(ctx, bson.M{"name": cins(args[2])}); err != nil {
				return err
			} else if c == 0 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Вомбата с таким ником не найдено...", update.Message.Chat.ID)
				return err
			}
			var (
				sClan Clan
				kWomb User
			)
			if err := clans.FindOne(ctx, bson.M{"leader": update.Message.From.ID}).Decode(&sClan); err != nil {
				return err
			}
			if err := users.FindOne(ctx, bson.M{"name": cins(args[2])}).Decode(&kWomb); err != nil {
				return err
			}
			var is bool
			for _, id := range sClan.Members {
				if id == kWomb.ID {
					is = true
					break
				}
			}
			if !is {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Вомбат с этим ником не состоит в Вашем клане",
					update.Message.Chat.ID,
				)
				return err
			}
			if kWomb.ID == update.Message.From.ID {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Если хотите выйти из клана, то напишите `клан выйти`",
					update.Message.Chat.ID,
				)
				return err
			}
			var appmsg string
			var nm []int64
			for _, id := range sClan.Members {
				if id == kWomb.ID {
					continue
				}
				nm = append(nm, id)
			}
			sClan.Members = nm
			if kWomb.ID == sClan.Banker {
				appmsg = "Теперь казначеем стали Вы."
				sClan.Banker = sClan.Leader
			}
			if err := docUpd(sClan, bson.M{"_id": sClan.Tag}, clans); err != nil {
				return err
			}
			_, err = bot.ReplyWithMessage(update.Message.MessageID, "Готово!\n"+appmsg, update.Message.Chat.ID)
			if err != nil {
				return err
			}
			_, err = bot.SendMessage(fmt.Sprintf("Вас кикнули из клана `%s` [%s]", sClan.Name, sClan.Tag), kWomb.ID)
			return err
		},
	},
	{
		Name: "ban",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "бан"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if len(args) == 2 {
				_, err := bot.ReplyWithMessage(update.Message.MessageID, "кого?", update.Message.Chat.ID)
				return err
			}
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if !isInUsers {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Кланы — приватная территория вомбатов. У тебя вомбата нет.",
					update.Message.Chat.ID,
				)
				return err
			}
			if c, err := clans.CountDocuments(ctx, bson.M{"leader": update.Message.From.ID}); err != nil {
				return err
			} else if c == 0 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Вы не являетесь лидером ни одного клана", update.Message.Chat.ID)
				return err
			}
			if c, err := users.CountDocuments(ctx, bson.M{"name": cins(args[2])}); err != nil {
				return err
			} else if c == 0 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Вомбата с таким ником не найдено...", update.Message.Chat.ID)
				return err
			}
			var (
				sClan Clan
				kWomb User
			)
			if err := clans.FindOne(ctx, bson.M{"leader": update.Message.From.ID}).Decode(&sClan); err != nil {
				return err
			}
			if err := users.FindOne(ctx, bson.M{"name": cins(args[2])}).Decode(&kWomb); err != nil {
				return err
			}
			var is bool
			for _, id := range sClan.Members {
				if id == kWomb.ID {
					is = true
					break
				}
			}
			if !is {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Вомбат с этим ником не состоит в Вашем клане", update.Message.Chat.ID)
				return err
			}
			if kWomb.ID == update.Message.From.ID {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Если Вы хотите быть забанеными, то передайте права лидера и попросите забанить Вас нового лидера",
					update.Message.Chat.ID,
				)
				return err
			}
			for _, id := range sClan.Banned {
				if id == kWomb.ID {
					_, err = bot.ReplyWithMessage(update.Message.MessageID, "Этот вомбат уже забанен", update.Message.Chat.ID)
					return err
				}
			}
			sClan.Banned = append(sClan.Banned, kWomb.ID)
			var appmsg string
			var nm []int64
			for _, id := range sClan.Members {
				if id == kWomb.ID {
					continue
				}
				nm = append(nm, id)
			}
			sClan.Members = nm
			if kWomb.ID == sClan.Banker {
				appmsg = "Теперь казначеем стали Вы."
				sClan.Banker = sClan.Leader
			}
			if err := docUpd(sClan, bson.M{"_id": sClan.Tag}, clans); err != nil {
				return err
			}
			_, err = bot.ReplyWithMessage(update.Message.MessageID, "Готово!\n"+appmsg, update.Message.Chat.ID)
			if err != nil {
				return err
			}
			_, err = bot.SendMessage(fmt.Sprintf("Вас забанили в клане `%s` [%s]", sClan.Name, sClan.Tag), kWomb.ID)
			return err
		},
	},
	{
		Name: "unban",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "разбан"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if len(args) == 2 {
				_, err := bot.ReplyWithMessage(update.Message.MessageID, "кого?", update.Message.Chat.ID)
				return err
			}
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if !isInUsers {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Кланы — приватная территория вомбатов. У тебя вомбата нет.",
					update.Message.Chat.ID,
				)
				return err
			}
			if c, err := clans.CountDocuments(ctx, bson.M{"leader": update.Message.From.ID}); err != nil {
				return err
			} else if c == 0 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Вы не являетесь лидером ни одного клана", update.Message.Chat.ID)
				return err
			}
			if c, err := users.CountDocuments(ctx, bson.M{"name": cins(args[2])}); err != nil {
				return err
			} else if c == 0 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Вомбата с таким ником не найдено...", update.Message.Chat.ID)
				return err
			}
			var (
				sClan Clan
				kWomb User
			)
			if err := clans.FindOne(ctx, bson.M{"leader": update.Message.From.ID}).Decode(&sClan); err != nil {
				return err
			}
			if err := users.FindOne(ctx, bson.M{"name": cins(args[2])}).Decode(&kWomb); err != nil {
				return err
			}
			var is bool
			var nb []int64
			for _, id := range sClan.Banned {
				if id == kWomb.ID {
					is = true
					continue
				}
				nb = append(nb, id)
			}
			if !is {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Данный вомбат не забанен в Вашем клане", update.Message.Chat.ID)
				return err
			}
			sClan.Banned = nb
			if err := docUpd(sClan, bson.M{"_id": sClan.Tag}, clans); err != nil {
				return err
			}
			_, err = bot.ReplyWithMessage(update.Message.MessageID, "Успешно!", update.Message.Chat.ID)
			if err != nil {
				return err
			}
			_, err = bot.SendMessage(fmt.Sprintf("Вы были разбанены в клане `%s` [%s]", sClan.Name, sClan.Tag), kWomb.ID)
			return err
		},
	},
	{
		Name: "rename",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "переименовать"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if len(args) < 3 {
				_, err := bot.ReplyWithMessage(
					update.Message.MessageID,
					"Слишком мало аргументов! Синтаксис: `клан переименовать [имя (можно пробелы)]`",
					update.Message.Chat.ID,
				)
				return err
			}
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if !isInUsers {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Кланы — приватная территория вомбатов. У тебя вомбата нет.",
					update.Message.Chat.ID,
				)
				return err
			}
			if c, err := clans.CountDocuments(ctx, bson.M{"leader": update.Message.From.ID}); err != nil {
				return err
			} else if c == 0 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Вы не являетесь лидером ни в одном клане",
					update.Message.Chat.ID,
				)
				return err
			}
			name := strings.Join(args[2:], " ")
			if len([]rune(name)) > 64 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Слишком длинное имя! Оно должно быть максимум 64 символов",
					update.Message.Chat.ID,
				)
				return err
			} else if len([]rune(name)) < 2 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Слишком короткое имя! Оно должно быть минимум 3 символа",
					update.Message.Chat.ID,
				)
				return err
			}
			if _, err := clans.UpdateOne(ctx, bson.M{"leader": update.Message.From.ID}, bson.M{
				"$set": bson.M{
					"name": name,
				},
			}); err != nil {
				return err
			}
			_, err = bot.ReplyWithMessage(
				update.Message.MessageID,
				fmt.Sprintf("Имя Вашего клана было успешно сменено на `%s`", name),
				update.Message.Chat.ID,
			)
			return err
		},
	},
	{
		Name: "settings",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "настройки"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if len(args) > 4 {
				_, err := bot.ReplyWithMessage(update.Message.MessageID, "слишком много аргументов", update.Message.Chat.ID)
				return err
			}
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if !isInUsers {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Кланы — приватная территория вомбатов. У тебя вомбата нет.",
					update.Message.Chat.ID,
				)
				return err
			}
			if c, err := clans.CountDocuments(ctx, bson.M{"leader": update.Message.From.ID}); err != nil {
				return err
			} else if c == 0 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Вы не являетесь лидером ни одного клана.",
					update.Message.Chat.ID,
				)
				return err
			}
			var sClan Clan
			if err := clans.FindOne(ctx, bson.M{"leader": update.Message.From.ID}).Decode(&sClan); err != nil {
				return err
			}
			if len(args) == 2 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					fmt.Sprintf(
						"Настройки клана:\n"+
							"  доступен_для_входа: %s",
						bool2string(sClan.Settings.AviableToJoin),
					),
					update.Message.Chat.ID,
				)
				return err
			}
			switch strings.ToLower(args[2]) {
			case "доступен_для_входа":
				if len(args) == 3 {
					_, err = bot.ReplyWithMessage(
						update.Message.MessageID,
						"доступен_для_входа: "+bool2string(sClan.Settings.AviableToJoin),
						update.Message.Chat.ID,
					)
					return err
				} else if ans := strings.ToLower(args[3]); ans == "да" {
					sClan.Settings.AviableToJoin = true
				} else if ans == "нет" {
					sClan.Settings.AviableToJoin = false
				} else {
					_, err = bot.ReplyWithMessage(
						update.Message.MessageID,
						"Поддерживаются только ответы `да` и `нет`",
						update.Message.Chat.ID,
					)
					return err
				}
			default:
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Настройка с такимименем не обнаружена",
					update.Message.Chat.ID,
				)
				return err
			}
			if err := docUpd(sClan, bson.M{"leader": update.Message.From.ID}, clans); err != nil {
				return err
			}
			_, err = bot.ReplyWithMessage(
				update.Message.MessageID,
				fmt.Sprintf(
					"Настройка `%s` теперь имеет значение `%s`",
					strings.ToLower(args[2]),
					strings.ToLower(args[3]),
				),
				update.Message.Chat.ID,
			)
			return err
		},
	},
	{
		Name: "bank",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "казна"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if len(args) == 2 {
				_, err := bot.ReplyWithMessage(update.Message.MessageID, "жесь", update.Message.Chat.ID)
				return err
			}
			for _, cmd := range clanBankCommands {
				if cmd.Is(args, update) {
					err := cmd.Action(args, update, womb)
					if err != nil {
						err = fmt.Errorf("%s: %v", cmd.Name, err)
					}
					return err
				}
			}
			_, err := bot.ReplyWithMessage(update.Message.MessageID, "не знаю такой команды", update.Message.Chat.ID)
			return err
		},
	},
	{
		Name: "attack",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "атака"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if len(args) == 2 {
				_, err := bot.ReplyWithMessage(update.Message.MessageID, "ихецац", update.Message.Chat.ID)
				return err
			}
			for _, cmd := range clanAttackCommands {
				if cmd.Is(args, update) {
					err := cmd.Action(args, update, womb)
					if err != nil {
						err = fmt.Errorf("%s: %v", cmd.Name, err)
					}
					return err
				}
			}
			_, err := bot.ReplyWithMessage(update.Message.MessageID, "не знаю такой команды", update.Message.Chat.ID)
			return err
		},
	},
}

var clanBankCommands = []command{
	{
		Name: "bank",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[2]) == "казна"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			_, err := bot.ReplyWithMessage(
				update.Message.MessageID,
				strings.Repeat("казна ", 42),
				update.Message.Chat.ID,
			)
			return err
		},
	},
	{
		Name: "take",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[2]) == "снять"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if len(args) != 4 {
				_, err := bot.ReplyWithMessage(
					update.Message.MessageID,
					"Слишком мало или много аргументов",
					update.Message.Chat.ID,
				)
				return err
			}
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if !isInUsers {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Кланы — приватная территория вомбатов. У тебя вомбата нет.",
					update.Message.Chat.ID,
				)
				return err
			}
			if c, err := clans.CountDocuments(ctx, bson.M{"members": womb.ID}); err != nil {
				return err
			} else if c == 0 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Вы не состоите ни в одном клане", update.Message.Chat.ID)
				return err
			}
			var sClan Clan
			if err := clans.FindOne(ctx, bson.M{"members": womb.ID}).Decode(&sClan); err != nil {
				return err
			}
			if !(sClan.Leader == womb.ID || sClan.Banker == womb.ID) {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Ошибка: вы не обладаете правом снимать деньги с казны (только лидер и казначей)",
					update.Message.Chat.ID,
				)
				return err
			}
			var take uint64
			if take, err = strconv.ParseUint(args[3], 10, 64); err != nil {
				if args[3] == "всё" {
					take = uint64(sClan.Money)
				} else {
					_, err = bot.ReplyWithMessage(
						update.Message.MessageID,
						"Ошибка: введено не число, либо число больше 2^63, либо отрицательное, либо дробное. короче да.",
						update.Message.Chat.ID,
					)
					return err
				}
			}
			if take > uint64(sClan.Money) {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "Запрашиваемая сумма выше количества денег в казне", update.Message.Chat.ID)
				return err
			} else if take == 0 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Хитр(ый/ая) как(ой/ая)",
					update.Message.Chat.ID,
				)
				return err
			}
			if _, err = clans.UpdateOne(ctx, bson.M{"_id": sClan.Tag},
				bson.M{"$inc": bson.M{"money": -int(take)}}); err != nil {
				return err
			} else if _, err = users.UpdateOne(ctx, bson.M{"_id": womb.ID},
				bson.M{"$inc": bson.M{"money": int(take)}}); err != nil {
			}
			_, err = bot.ReplyWithMessage(
				update.Message.MessageID,
				fmt.Sprintf(
					"Вы успешно сняли из казны %d Ш! Теперь в казне %d Ш, а у вас на счету %d",
					take, uint64(sClan.Money)-take, uint64(womb.Money)+take,
				),
				update.Message.Chat.ID,
			)
			return err
		},
	},
	{
		Name: "put",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[2]) == "положить"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if len(args) != 4 {
				_, err := bot.ReplyWithMessage(
					update.Message.MessageID,
					"Слишком много или мало аргументов",
					update.Message.Chat.ID,
				)
				return err
			}
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if !isInUsers {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Кланы — приватная территория вомбатов. У тебя вомбата нет.",
					update.Message.Chat.ID,
				)
				return err
			}
			if c, err := clans.CountDocuments(ctx, bson.M{"members": update.Message.From.ID}); err != nil {
				return err
			} else if c == 0 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Вы не состоите ни в одном клане",
					update.Message.Chat.ID,
				)
				return err
			}
			var sClan Clan
			if err := clans.FindOne(ctx, bson.M{"members": update.Message.From.ID}).Decode(&sClan); err != nil {
				return err
			}
			var (
				take uint64
			)
			if take, err = strconv.ParseUint(args[3], 10, 64); err != nil {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Ошибка: введено не число, либо число больше 2^63, либо отрицательное, либо дробное. короче да.",
					update.Message.Chat.ID,
				)
				return err
			} else if take > uint64(womb.Money) {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Сумма, которую вы хотите положить, больше кол-ва денег на вашем счету",
					update.Message.Chat.ID,
				)
				return err
			} else if take == 0 {
				_, err = bot.ReplyWithMessage(update.Message.MessageID, "блин", update.Message.Chat.ID)
				return err
			}
			if _, err := users.UpdateOne(ctx, bson.M{"_id": womb.ID}, bson.M{
				"$inc": bson.M{
					"money": -int(take),
				},
			}); err != nil {
				return err
			} else if _, err := clans.UpdateOne(ctx, bson.M{"_id": sClan.Tag}, bson.M{
				"$inc": bson.M{
					"money": int(take),
				},
			}); err != nil {
				return err
			}
			_, err = bot.ReplyWithMessage(
				update.Message.MessageID,
				fmt.Sprintf(
					"Вы положили %d Ш в казну. Теперь в казне %d Ш, а у вас %d",
					take, uint64(sClan.Money)+take, uint64(womb.Money)-take,
				),
				update.Message.Chat.ID,
			)
			if sClan.Leader != womb.ID {
				_, lerr := bot.SendMessage(
					fmt.Sprintf("%s положил(а) %d шишей в казну клана", womb.Name, take),
					sClan.Leader,
				)
				if lerr != nil {
					return lerr
				}
			}
			if sClan.GroupID != update.Message.Chat.ID {
				_, gerr := bot.SendMessage(
					fmt.Sprintf("%s положил(а) %d шишей в казну клана", womb.Name, take),
					sClan.GroupID,
				)
				if gerr != nil {
					return gerr
				}
			}
			return err
		},
	},
}

var clanAttackCommands = []command{
	{
		Name: "attack",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[2]) == "атака"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			_, err := bot.ReplyWithMessage(
				update.Message.MessageID,
				strings.Repeat("атака ", 42),
				update.Message.Chat.ID,
			)
			return err
		},
	},
	{
		Name: "to",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[2]) == "на"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if len(args) == 3 {
				_, err := bot.ReplyWithMessage(
					update.Message.MessageID,
					"Атака на: на кого?",
					update.Message.Chat.ID,
				)
				return err
			} else if len(args) > 4 {
				_, err := bot.ReplyWithMessage(
					update.Message.MessageID,
					"Атака на: слишком много аргументов",
					update.Message.Chat.ID,
				)
				return err
			}
			var err error
			if rCount, err := clans.CountDocuments(ctx,
				bson.M{"members": update.Message.From.ID}); err != nil {
				return err
			} else if rCount == 0 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Вы не состоите ни в одном клане",
					update.Message.Chat.ID,
				)
				return err
			} else if rCount, err := clans.CountDocuments(ctx,
				bson.M{"leader": update.Message.From.ID}); err != nil {
				return err
			} else if rCount == 0 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Вы не являетесь лидером клана, в котором состоите",
					update.Message.Chat.ID,
				)
				return err
			}
			var fromClan Clan
			err = clans.FindOne(ctx, bson.M{"leader": update.Message.From.ID}).Decode(&fromClan)
			if err != nil {
				return err
			} else if ok, from := isInClattacks(fromClan.Tag, clattacks); from {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Вы уже нападаете на другой клан",
					update.Message.Chat.ID,
				)
				return err
			} else if ok {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"На вас уже нападают)",
					update.Message.Chat.ID,
				)
				return err
			}
			tag := strings.ToUpper(args[3])
			if len([]rune(tag)) > 64 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Ошибка: слишком длинный тег!",
					update.Message.Chat.ID,
				)
				return err
			} else if !isValidTag(tag) {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Нелегальный тег",
					update.Message.Chat.ID,
				)
				return err
			} else if fromClan.Tag == tag {
				bot.ReplyWithMessage(
					update.Message.MessageID,
					"гений",
					update.Message.Chat.ID,
				)
				return err
			} else if rCount, err := clans.CountDocuments(ctx,
				bson.M{"_id": tag}); err != nil {
				return err
			} else if rCount == 0 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Ошибка: клана с таким тегом не найдено",
					update.Message.Chat.ID,
				)
				return err
			} else if ok, from := isInClattacks(tag, clattacks); from {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Клан ["+tag+"] уже атакует кого-то",
					update.Message.Chat.ID,
				)
				return err
			} else if ok {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Клан ["+tag+"] уже атакуется",
					update.Message.Chat.ID,
				)
				return err
			}
			var toClan Clan
			err = clans.FindOne(ctx, bson.M{"_id": tag}).Decode(&toClan)
			if err != nil {
				return err
			}
			newClat := Clattack{
				ID:   fromClan.Tag + "_" + tag,
				From: fromClan.Tag,
				To:   tag,
			}
			_, err = clattacks.InsertOne(ctx, newClat)
			if err != nil {
				return err
			}
			_, err = bot.ReplyWithMessage(
				update.Message.MessageID,
				"Отлично! Вы отправили вомбатов ждать согласия на вомбой",
				update.Message.Chat.ID,
			)
			if err != nil {
				return err
			}
			_, err = bot.SendMessage(
				"АААА!!! НА ВАС НАПАЛ КЛАН "+fromClan.Tag+". предпримите что-нибудь(",
				toClan.Leader,
			)
			return err
		},
	},
	{
		Name: "cancel",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[2]) == "отмена"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if len(args) != 3 {
				_, err := bot.ReplyWithMessage(
					update.Message.MessageID,
					"Клан атака отмена: слишком много аргументов",
					update.Message.Chat.ID,
				)
				return err
			}
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if !isInUsers {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Кланы — приватная территория вомбатов. У тебя вомбата нет.",
					update.Message.Chat.ID,
				)
				return err
			}
			if rCount, err := clans.CountDocuments(ctx,
				bson.M{"members": update.Message.From.ID}); err != nil {
				return err
			} else if rCount == 0 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Ошибка: вы не состоите ни в одном клане",
					update.Message.Chat.ID,
				)
				return err
			} else if rCount, err := clans.CountDocuments(ctx,
				bson.M{"leader": update.Message.From.ID}); err != nil {
				return err
			} else if rCount == 0 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Ошибка: вы не являетесь лидером в своём клане",
					update.Message.Chat.ID,
				)
				return err
			}
			var cClan Clan
			err = clans.FindOne(ctx, bson.M{"leader": update.Message.From.ID}).Decode(&cClan)
			if err != nil {
				return err
			}
			is, isfr := isInClattacks(cClan.Tag, clattacks)
			if !is {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Вы никого не атакуете и никем не атакуетесь. Вам нечего отменять :)",
					update.Message.Chat.ID)
				return err
			}
			var clat Clattack
			err = clattacks.FindOne(ctx, bson.M{func(isfr bool) string {
				if isfr {
					return "from"
				}
				return "to"
			}(isfr): cClan.Tag}).Decode(&clat)
			if err != nil {
				return err
			}
			var (
				send  bool = true
				oClan Clan
			)
			if rCount, err := clans.CountDocuments(ctx,
				bson.M{"_id": func(clat Clattack, isfr bool) string {
					if isfr {
						return clat.To
					}
					return clat.From
				}(clat, isfr)}); err != nil {
				return err
			} else if rCount == 0 {
				send = false
			} else {
				err = clans.FindOne(ctx, bson.M{"_id": func(clat Clattack, isfr bool) string {
					if isfr {
						return clat.To
					}
					return clat.From
				}(clat, isfr)}).Decode(&oClan)
				if err != nil {
					return err
				}
			}
			_, err = clattacks.DeleteOne(ctx, bson.M{"to": clat.To})
			if err != nil {
				return err
			}
			can0, err := getImgs(imgsC, "cancel_0")
			if err != nil {
				return err
			}
			var can1 Imgs
			if send {
				can1, err = getImgs(imgsC, "cancel_1")
				if err != nil {
					return err
				}
			}
			_, err = bot.ReplyWithPhoto(
				update.Message.MessageID, randImg(can0), "Вы "+func(isfr bool) string {
					if isfr {
						return "отменили"
					}
					return "отклонили"
				}(isfr)+" клановую атаку",
				update.Message.Chat.ID,
			)
			if send {
				_, err = bot.SendPhoto(
					randImg(can1),
					"Вашу клановую атаку "+func(isfr bool) string {
						if isfr {
							return "отменили"
						}
						return "отклонили"
					}(isfr)+")",
					oClan.Leader,
				)
				if err != nil {
					return err
				}
			}
			return err
		},
	},
	{
		Name: "accept",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[2]) == "принять"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if len(args) != 3 {
				_, err := bot.ReplyWithMessage(
					update.Message.MessageID,
					"Слишком много аргументов",
					update.Message.Chat.ID,
				)
				return err
			}
			var err error
			if rCount, err := clans.CountDocuments(ctx,
				bson.M{"leader": update.Message.From.ID}); err != nil {
				return err
			} else if rCount == 0 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Вы не являетесь лидером ни одного клана",
					update.Message.Chat.ID,
				)
				return err
			}
			var toClan Clan
			err = clans.FindOne(ctx, bson.M{"leader": update.Message.From.ID}).Decode(&toClan)
			if err != nil {
				return err
			}
			if is, isfr := isInClattacks(toClan.Tag, clattacks); !is {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Ваш клан не атакуется/не атакует",
					update.Message.Chat.ID,
				)
				return err
			} else if isfr {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Принимать вомбой может только атакуемая сторона",
					update.Message.Chat.ID,
				)
				return err
			}
			var clat Clattack
			err = clattacks.FindOne(ctx, bson.M{"to": toClan.Tag}).Decode(&clat)
			if err != nil {
				return err
			}
			if rCount, err := clans.CountDocuments(ctx, bson.M{"_id": clat.From}); err != nil {
				return err
			} else if rCount == 0 {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Ошибка: атакующего клана не существует!",
					update.Message.Chat.ID,
				)
				return err
			}
			var frClan Clan
			err = clans.FindOne(ctx, bson.M{"_id": clat.From}).Decode(&frClan)
			if err != nil {
				return err
			}
			var (
				toclwar, frclwar Clwar
				tWomb            User
			)
			for sClan, clw := range map[*Clan]*Clwar{&toClan: &toclwar, &frClan: &frclwar} {
				var lost uint8 = 0
				for _, id := range sClan.Members {
					if rCount, err := users.CountDocuments(ctx,
						bson.M{"_id": id}); err != nil {
						return err
					} else if rCount == 0 {
						lost++
						continue
					} else {
						err = users.FindOne(ctx, bson.M{"_id": id}).Decode(&tWomb)
						if err != nil {
							return err
						}
						if id == sClan.Leader {
						}
						clw.Health += tWomb.Health
						clw.Force += tWomb.Force
					}
					if uint32(len(sClan.Members)-int(lost)) != 0 {
						clw.Health /= uint32(len(sClan.Members) - int(lost))
						clw.Force /= uint32(len(sClan.Members) - int(lost))
					} else {
						_, err = bot.ReplyWithMessage(
							update.Message.MessageID,
							"Ошибка: у клана ["+sClan.Tag+"] все вомбаты потеряны( ответьте командой /admin",
							update.Message.Chat.ID,
						)
						return err
					}
				}
			}
			atimgs, err := getImgs(imgsC, "attacks")
			if err != nil {
				return err
			}
			im := randImg(atimgs)
			ph1, err := bot.ReplyWithPhoto(update.Message.MessageID, im, "", update.Message.Chat.ID)
			if err != nil {
				return err
			}
			var frClanID int64 = frClan.Leader
			if update.Message.Chat.ID != frClan.GroupID {
				frClanID = frClan.GroupID
			}
			ph2, err := bot.SendPhoto(im, "", frClanID)
			if err != nil {
				return err
			}
			war1, err := bot.ReplyWithMessage(ph1, "Да начнётся вомбой!", update.Message.Chat.ID)
			if err != nil {
				return err
			}
			war2, err := bot.ReplyWithMessage(ph2, fmt.Sprintf(
				"АААА ВАЙНААААА!!!\n Вомбат %s всё же принял ваше предложение",
				womb.Name), frClanID,
			)
			if err != nil {
				return err
			}
			time.Sleep(5 * time.Second)
			h1, h2 := int(toclwar.Health), int(frclwar.Health)
			for _, round := range []int{1, 2, 3} {
				f1 := uint32(2 + rand.Intn(int(toclwar.Force-1)))
				f2 := uint32(2 + rand.Intn(int(frclwar.Force-1)))
				err = bot.EditMessage(
					war1, fmt.Sprintf(
						"РАУНД %d\n\n[%s]:\n - здоровье: %d\n - Ваш удар: %d\n\n[%s]:\n - здоровье: %d",
						round, toClan.Tag, h1, f1, frClan.Tag, h2),
					update.Message.Chat.ID,
				)
				if err != nil {
					return err
				}
				err = bot.EditMessage(war2, fmt.Sprintf(
					"РАУНД %d\n\n[%s]:\n - здоровье: %d\n - Ваш удар: %d\n\n[%s]:\n - здоровье: %d",
					round, frClan.Tag, h2, f2, toClan.Tag, h1), frClanID,
				)
				if err != nil {
					return err
				}
				time.Sleep(3 * time.Second)
				h1 -= int(f2)
				h2 -= int(f1)
				bot.EditMessage(war1, fmt.Sprintf(
					"РАУНД %d\n\n[%s]\n - здоровье: %d\n - Ваш удар: %d\n\n[%s]:\n - здоровье: %d\n - 💔 удар: %d",
					round, toClan.Tag, h1, f1, frClan.Tag, h2, f2), update.Message.Chat.ID,
				)
				if err != nil {
					return err
				}
				bot.EditMessage(war2, fmt.Sprintf(
					"РАУНД %d\n\n[%s]:\n - здоровье: %d\n - Ваш удар: %d\n\n[%s]:\n - здоровье: %d\n - 💔 удар: %d",
					round, frClan.Tag, h2, f2, toClan.Tag, h1, f1), frClanID,
				)
				if err != nil {
					return err
				}
				time.Sleep(5 * time.Second)
				if int(h2)-int(f1) <= 5 && int(h1)-int(f2) <= 5 {
					err = bot.EditMessage(war1,
						"Оба клана сдохли!!!)\nВаши характеристики не поменялись, но зато да.",
						update.Message.Chat.ID,
					)
					if err != nil {
						return err
					}
					err = bot.EditMessage(war2,
						"Оба клана сдохли!!!)\nВаши характеристики не поменялись, но зато да.",
						frClanID,
					)
					if err != nil {
						return err
					}

					time.Sleep(5 * time.Second)
					break
				} else if int(h2)-int(f1) <= 5 {
					err = bot.EditMessage(war1, fmt.Sprintf(
						"В раунде %d благодаря силе участников победил клан...",
						round), update.Message.Chat.ID,
					)
					if err != nil {
						return err
					}
					err = bot.EditMessage(war2, fmt.Sprintf(
						"В раунде %d благодаря лишению у другого здоровья победил клан...",
						round), frClanID,
					)
					if err != nil {
						return err
					}
					time.Sleep(3 * time.Second)
					toClan.XP += 10
					err = bot.EditMessage(war1, fmt.Sprintf(
						"Победил клан `%s` [%s]!!!\nВы получили 10 XP, теперь их у вас %d",
						toClan.Name, toClan.Tag, toClan.XP), update.Message.Chat.ID,
					)
					if err != nil {
						return err
					}
					err = bot.EditMessage(war2, fmt.Sprintf(
						"Победил клан `%s` [%s]!!!\nВаше состояние не изменилось)",
						toClan.Name, toClan.Tag), frClanID,
					)
					if err != nil {
						return err
					}
					break
				} else if int(h1)-int(f2) <= 5 {
					err = bot.EditMessage(war1, fmt.Sprintf(
						"В раунде %d благодаря силе участников победил клан...",
						round), update.Message.Chat.ID,
					)
					if err != nil {
						return err
					}
					err = bot.EditMessage(war2, fmt.Sprintf(
						"В раунде %d благодаря лишению у другого здоровья победил клан...",
						round), frClanID,
					)
					if err != nil {
						return err
					}
					time.Sleep(3 * time.Second)
					frClan.XP += 10
					err = bot.EditMessage(war2, fmt.Sprintf(
						"Победил клан `%s` %s!!!\nВы получили 10 XP, теперь их у Вас %d",
						frClan.Name, frClan.Tag, frClan.XP), frClanID,
					)
					if err != nil {
						return err
					}
					womb.Health = 5
					womb.Money = 50
					err = bot.EditMessage(war1, fmt.Sprintf(
						"Победил клан `%s` [%s]!!!\nВаше состояние не изменилось)",
						frClan.Name, frClan.Tag), update.Message.Chat.ID,
					)
					if err != nil {
						return err
					}

					break
				} else if round == 3 {
					frClan.XP += 10
					if h1 < h2 {
						err = bot.EditMessage(war2, fmt.Sprintf(
							"И победил клан `%s` %s!!!\nВы получили 10 XP, теперь их у Вас %d",
							frClan.Name, frClan.Tag, frClan.XP), frClanID,
						)
						if err != nil {
							return err
						}
						err = bot.EditMessage(war1, fmt.Sprintf(
							"И победил клан `%s` [%s]!!!\nВаше состояние не изменилось)",
							frClan.Name, frClan.Tag), update.Message.Chat.ID,
						)
						if err != nil {
							return err
						}
					} else {
						toClan.XP += 10
						err = bot.EditMessage(war1, fmt.Sprintf(
							"Победил клан `%s` [%s]!!!\nВы получили 10 XP, теперь их у вас %d",
							toClan.Name, toClan.Tag, toClan.XP), update.Message.Chat.ID,
						)
						if err != nil {
							return err
						}
						err = bot.EditMessage(war2, fmt.Sprintf(
							"Победил клан `%s` [%s]!!!\nВаше состояние не изменилось)",
							toClan.Name, toClan.Tag), frClanID,
						)
						if err != nil {
							return err
						}
					}
				}
			}
			err = docUpd(toClan, bson.M{"_id": toClan.Tag}, clans)
			if err != nil {
				return err
			}
			err = docUpd(frClan, bson.M{"_id": frClan.Tag}, clans)
			if err != nil {
				return err
			}
			_, err = clattacks.DeleteOne(ctx, bson.M{"_id": clat.ID})
			return err
		},
	},
	{
		Name: "status",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[2]) == "статус"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			var (
				sClan Clan
				err   error
			)
			switch len(args) {
			case 3:
				isInUsers, err := getIsInUsers(update.Message.From.ID)
				if !isInUsers {
					_, err = bot.ReplyWithMessage(
						update.Message.MessageID,
						"Вы не имеете вомбата => Вы не состоите ни в одном клане. Добавьте тег.",
						update.Message.Chat.ID,
					)
					return err
				}
				if c, err := clans.CountDocuments(ctx, bson.M{"members": update.Message.From.ID}); err != nil {
					return err
				} else if c == 0 {
					_, err = bot.ReplyWithMessage(
						update.Message.MessageID,
						"Вы не состоите ни в одном клане. Добавьте тег.",
						update.Message.Chat.ID,
					)
					return err
				}
				if err := clans.FindOne(ctx, bson.M{"members": update.Message.From.ID}).Decode(&sClan); err != nil {
					return err
				}
			case 4:
				tag := strings.ToUpper(args[3])
				if len(tag) < 3 || len(tag) > 5 {
					_, err = bot.ReplyWithMessage(update.Message.MessageID, "Некорректный тег", update.Message.Chat.ID)
					return err
				}
				if c, err := clans.CountDocuments(ctx, bson.M{"_id": tag}); err != nil {
					return err
				} else if c == 0 {
					_, err = bot.ReplyWithMessage(update.Message.MessageID, "Клана с таким тегом нет...", update.Message.Chat.ID)
					return err
				}
				if err := clans.FindOne(ctx, bson.M{"_id": tag}).Decode(&sClan); err != nil {
					return err
				}
			default:
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"СЛИШКОМ. МНОГО. АРГУМЕНТОВ(((",
					update.Message.Chat.ID,
				)
				return err
			}
			var (
				is            bool
				isfr          bool
				sClanPosition string = "to"
			)
			if is, isfr = isInClattacks(sClan.Tag, clattacks); !is {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"Этот клан не учавствует в атаках)",
					update.Message.Chat.ID,
				)
				return err
			}
			if isfr {
				sClanPosition = "from"
			}
			var (
				sClat Clattack
			)
			if err := clattacks.FindOne(ctx, bson.M{
				sClanPosition: sClan.Tag,
			}).Decode(&sClat); err != nil {
				return err
			}
			_, err = bot.ReplyWithMessage(
				update.Message.MessageID,
				fmt.Sprintf(
					"От: [%s]\nНа: [%s]",
					sClat.From,
					sClat.To,
				),
				update.Message.Chat.ID,
			)
			return err
		},
	},
}

var devtoolsCommands = []command{
	{
		Name: "set_money",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "set_money"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if !isInUsers {
				return nil
			}
			if len(args) < 3 {
				_, err := bot.ReplyWithMessage(
					update.Message.MessageID,
					"мало аргументов",
					update.Message.Chat.ID,
				)
				return err
			}
			var i uint64
			if i, err = strconv.ParseUint(args[2], 10, 32); err != nil {
				_, err = bot.ReplyWithMessage(
					update.Message.MessageID,
					"не число",
					update.Message.Chat.ID,
				)
				return err
			}
			_, err = users.UpdateOne(ctx, bson.M{"_id": womb.ID}, bson.M{"$set": bson.M{"money": i}})
			if err != nil {
				debl.Println(err)
				return err
			}
			_, err = bot.ReplyWithMessage(
				update.Message.MessageID,
				"успешно",
				update.Message.Chat.ID,
			)
			return err
		},
	},
	{
		Name: "reset",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "reset"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if len(args) < 3 {
				_, err := bot.ReplyWithMessage(
					update.Message.MessageID,
					"мало аргументов",
					update.Message.Chat.ID,
				)
				return err
			}
			switch strings.ToLower(args[2]) {
			case "force":
				womb.Force = 2
			case "health":
				womb.Health = 5
			case "xp":
				womb.XP = 0
			case "all":
				womb.Force = 2
				womb.Health = 5
				womb.XP = 0
			default:
				_, err := bot.ReplyWithMessage(
					update.Message.MessageID,
					"режимы: force/health/xp/all",
					update.Message.Chat.ID,
				)
				return err
			}
			err := docUpd(womb, wombFilter(womb), users)
			if err != nil {
				return err
			}
			_, err = bot.ReplyWithMessage(
				update.Message.MessageID,
				"успешно",
				update.Message.Chat.ID,
			)
			return err
		},
	},
	{
		Name: "info",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "info"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			var sWomb User
			if len(args) == 3 {
				if c, err := users.CountDocuments(ctx, bson.M{"name": cins(args[2])}); err != nil {
					return err
				} else if c == 0 {
					_, err = bot.ReplyWithMessage(
						update.Message.MessageID,
						"нет такого/такой",
						update.Message.Chat.ID,
					)
					return err
				}
				err := users.FindOne(ctx, bson.M{"name": cins(args[2])}).Decode(&sWomb)
				if err != nil {
					return err
				}
			} else {
				sWomb = womb
			}
			_, err := bot.ReplyWithMessage(
				update.Message.MessageID,
				fmt.Sprintf(
					"%#v", sWomb,
				),
				update.Message.Chat.ID,
			)
			return err
		},
	},
	{
		Name: "help",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "help"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			_, err := bot.ReplyWithMessage(
				update.Message.MessageID,
				"https://telegra.ph/Vombot-devtools-help-10-28",
				update.Message.Chat.ID,
			)
			return err
		},
	},
	{
		Name: "invite",
		Is: func(args []string, update tg.Update) bool {
			return strings.ToLower(args[1]) == "invite"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if len(args) == 2 {
				return nil
			}
			_, err := users.UpdateOne(ctx, bson.M{"name": cins(args[2])}, bson.M{"$push": bson.M{"titles": 0}})
			if err != nil {
				return err
			}
			_, err = bot.ReplyWithMessage(
				update.Message.MessageID,
				"успешно",
				update.Message.Chat.ID,
			)
			return err
		},
	},
}
