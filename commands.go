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
	{
		Name: "change_name",
		Is: func(args []string, update tg.Update) bool {
			return strings.HasPrefix(
				strings.ToLower(strings.Join(args, " ")),
				"сменить имя",
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
				_, err = replyToMsg(update.Message.MessageID, "Да блин нафиг, вы вобмата забыли завести!!!!!!!", update.Message.From.ID, bot)
				return err
			} else if len(args) != 3 {
				if len(args) == 2 {
					_, err = replyToMsg(update.Message.MessageID, "вомбату нужно имя! ты его не указал", update.Message.From.ID, bot)
				} else {
					_, err = replyToMsg(update.Message.MessageID, "слишком много аргументов...", update.Message.From.ID, bot)
				}
				return err
			} else if hasTitle(1, womb.Titles) {
				_, err = replyToMsg(update.Message.MessageID, "Тебе нельзя, ты спамер (оспорить: /admin)", update.Message.From.ID, bot)
				return err
			} else if womb.Money < 3 {
				_, err = replyToMsg(update.Message.MessageID, "Мало шишей блин нафиг!!!!", update.Message.From.ID, bot)
				return err
			}
			name := args[2]
			if womb.Name == name {
				_, err = replyToMsg(update.Message.MessageID, "зачем", update.Message.From.ID, bot)
				return err
			} else if len([]rune(name)) > 64 {
				_, err = replyToMsg(update.Message.MessageID, "Слишком длинный никнейм!", update.Message.From.ID, bot)
				return err
			} else if isInList(name, []string{"вoмбoт", "вoмбoт", "вомбoт", "вомбот", "бот", "bot", "бoт", "bоt",
				"авто", "auto"}) {
				_, err = replyToMsg(update.Message.MessageID, "Такие никнеймы заводить нельзя", update.Message.From.ID, bot)
				return err
			} else if !isValidName(name) {
				_, err = replyToMsg(update.Message.MessageID, "Нелегальное имя:(\n", update.Message.From.ID, bot)
				return err
			}
			rCount, err := users.CountDocuments(ctx, bson.M{"name": cins(name)})
			if err != nil {
				return err
			} else if rCount != 0 {
				_, err = replyToMsg(update.Message.MessageID, fmt.Sprintf("Никнейм `%s` уже занят(", name), update.Message.From.ID, bot)
				return err
			}
			womb.Money -= 3
			caseName := strings.Join(args[2:], " ")
			womb.Name = caseName
			err = docUpd(womb, bson.M{"_id": update.Message.From.ID}, users)
			if err != nil {
				return err
			}
			_, err = replyToMsg(update.Message.MessageID,
				fmt.Sprintf("Теперь вашего вомбата зовут %s. С вашего счёта сняли 3 шиша", caseName),
				update.Message.From.ID, bot,
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
				_, err = replyToMsg(update.Message.MessageID, "А ты куда? У тебя вомбата нет...", update.Message.From.ID, bot)
				return err
			}

			if womb.Money < 1 {
				_, err = replyToMsg(update.Message.MessageID, "Охранники тебя прогнали; они требуют шиш за проход, а у тебя ни шиша нет", update.Message.From.ID, bot)
				return err
			}
			womb.Money--
			rand.Seed(time.Now().UnixNano())
			if ch := rand.Int(); ch%2 == 0 || hasTitle(2, womb.Titles) && (ch%2 == 0 || ch%3 == 0) {
				rand.Seed(time.Now().UnixNano())
				win := rand.Intn(9) + 1
				womb.Money += uint64(win)
				if addXP := rand.Intn(512 - 1); addXP < 5 {
					womb.XP += uint32(addXP)
					_, err = replyToMsg(update.Message.MessageID,
						fmt.Sprintf(
							"Поздравляем! Вы нашли на дороге %d шишей, а ещё вам дали %d XP! Теперь у вас %d шишей при себе и %d XP",
							win, addXP, womb.Money, womb.XP,
						),
						update.Message.From.ID, bot,
					)
				} else {
					_, err = replyToMsg(update.Message.MessageID,
						fmt.Sprintf(
							"Поздравляем! Вы нашли на дороге %d шишей! Теперь их у вас при себе %d", win, womb.Money,
						),
						update.Message.From.ID, bot,
					)
				}
				return err
			} else {
				_, err = replyToMsg(
					update.Message.MessageID, "Вы заплатили один шиш охранникам денежной дорожки, но увы, вы так ничего и не нашли",
					update.Message.From.ID, bot,
				)
				return err
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
			_, err := replyToMsg(update.Message.MessageID, strings.Join([]string{"Магазин:", " — 1 здоровье — 5 ш", " — 1 мощь — 3 ш",
				" — квес — 256 ш", " — вадшам — 250'000 ш",
				"Для покупки использовать команду 'купить [название_объекта] ([кол-во])",
			}, "\n"),
				update.Message.From.ID, bot,
			)
			return err
		},
	},
	{
		Name: "buy",
		Is: func(args []string, update tg.Update) bool {
			return args[0] == "купить"
		},
		Action: func(args []string, update tg.Update, womb User) error {
			if len(args) == 1 {
				_, err := replyToMsg(update.Message.MessageID, "купить", update.Message.Chat.ID, bot)
				return err
			}
			isInUsers, err := getIsInUsers(update.Message.From.ID)
			if err != nil {
				return err
			}
			if !isInUsers {
				_, err = replyToMsg(update.Message.MessageID, "у тебя недостаточно вомбатов чтобы кумпить (нужен минимум один)", update.Message.Chat.ID, bot)
				return err
			}
			switch args[1] {
			case "здоровья":
				fallthrough
			case "здоровье":
				if len(args) > 3 {
					_, err := replyToMsg(update.Message.MessageID, "Ошибка: слишком много аргументов...", update.Message.Chat.ID, bot)
					return err
				}
				var amount uint32 = 1
				if len(args) == 3 {
					if val, err := strconv.ParseUint(args[2], 10, 32); err == nil {
						if val == 0 {
							_, err = replyToMsg(update.Message.MessageID, "Поздравляю! Теперь у вас одна шиза и ещё одна шиза", update.Message.Chat.ID, bot)
							return err
						}
						amount = uint32(val)
					} else {
						_, err = replyToMsg(update.Message.MessageID, "Ошибка: число должно быть неотрицательным, целым и меньше 2^32", update.Message.Chat.ID, bot)
						return err
					}
				}
				if womb.Money >= uint64(amount*5) {
					if uint64(womb.Health+amount) < uint64(math.Pow(2, 32)) {
						womb.Money -= uint64(amount * 5)
						womb.Health += amount
						err = docUpd(womb, wombFilter(womb), users)
						if err != nil {
							return err
						}
						_, err = replyToMsg(update.Message.MessageID,
							fmt.Sprintf("Поздравляю! Теперь у вас %d здоровья и %d шишей при себе", womb.Health, womb.Money),
							update.Message.Chat.ID, bot,
						)
						return err
					} else {
						_, err = replyToMsg(update.Message.MessageID,
							"Ошибка: вы достигли максимального количества здоровья (2 в 32 степени). Если это вас возмущает, ответьте командой /admin",
							update.Message.Chat.ID, bot,
						)
						return err
					}
				} else {
					_, err = replyToMsg(update.Message.MessageID, "Надо накопить побольше шишей! 1 здоровье = 5 шишей", update.Message.Chat.ID, bot)
					return err
				}
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
					_, err = replyToMsg(update.Message.MessageID, "Ошибка: слишком много аргументов...", update.Message.Chat.ID, bot)
					return err
				}
				var amount uint32 = 1
				if len(args) == 3 {
					if val, err := strconv.ParseUint(args[2], 10, 32); err == nil {
						if val == 0 {
							_, err = replyToMsg(update.Message.MessageID, "Поздравляю! Теперь у вас одна шиза и ещё одна шиза", update.Message.Chat.ID, bot)
							return err
						}
						amount = uint32(val)
					} else {
						_, err = replyToMsg(update.Message.MessageID, "Ошибка: число должно быть неотрицательным, целым и меньше 2^32", update.Message.Chat.ID, bot)
						return err
					}
				}
				if womb.Money < uint64(amount*3) {
					_, err = replyToMsg(update.Message.MessageID,
						"Ошибка: вы достигли максимального количества мощи (2 в 32 степени). Если это вас возмущает, ответьте командой /admin",
						update.Message.Chat.ID, bot,
					)
					return err
				}
				if uint64(womb.Force+1) < uint64(math.Pow(2, 32)) {
					_, err = replyToMsg(update.Message.MessageID, "Надо накопить побольше шишей! 1 мощь = 3 шиша", update.Message.Chat.ID, bot)
					return err
				}
				womb.Money -= uint64(amount * 3)
				womb.Force += amount
				err = docUpd(womb, wombFilter(womb), users)
				if err != nil {
					return err
				}
				_, err = replyToMsg(update.Message.MessageID,
					fmt.Sprintf("Поздравляю! Теперь у вас %d мощи и %d шишей при себе", womb.Force, womb.Money),
					update.Message.Chat.ID, bot,
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
					_, err = replyToMsg(update.Message.MessageID, "ужас !! слишком много аргументов!!!", update.Message.From.ID, bot)
					return err
				} else if hasTitle(4, womb.Titles) {
					_, err = replyToMsg(update.Message.MessageID, "у вас уже есть вадшам", update.Message.From.ID, bot)
					return err
				} else if womb.Money < 250005 {
					_, err = replyToMsg(update.Message.MessageID, "Ошибка: недостаточно шишей для покупки (требуется 250000 + 5)", update.Message.From.ID, bot)
					return err
				}
				womb.Money -= 250000
				womb.Titles = append(womb.Titles, 4)
				err = docUpd(womb, wombFilter(womb), users)
				if err != nil {
					return err
				}
				_, err = replyToMsg(update.Message.MessageID, "Теперь вы вадшамообладатель", update.Message.From.ID, bot)
			case "квес":
				fallthrough
			case "квеса":
				fallthrough
			case "квесу":
				fallthrough
			case "qwess":
				if len(args) != 2 {
					_, err = replyToMsg(update.Message.MessageID, "Слишком много аргументов!", update.Message.From.ID, bot)
					return err
				} else if womb.Money < 256 {
					leps, err := getImgs(imgsC, "leps")
					if err != nil {
						return err
					}
					_, err = replyWithPhoto(update.Message.MessageID,
						randImg(leps),
						"Вы подошли к ближайшему ларьку, но, увы, кролик-Лепс на кассе сказал, что надо 256 шишей, а у вас, к сожалению, меньше",
						update.Message.From.ID, bot,
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
					_, err = replyWithPhoto(update.Message.MessageID,
						randImg(qwess),
						"Вы купили чудесного вкуса квес у кролика-Лепса в ларьке за 256 шишей. Глотнув этот напиток, вы поняли, что получили новый титул с ID 2",
						update.Message.From.ID, bot,
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
					_, err = replyWithPhoto(update.Message.MessageID,
						randImg(qwess),
						"Вы вновь купили вкусного квеса у того же кролика-Лепса в том же ларьке за 256 шишей. \"Он так освежает, я чувствую себя человеком\" — думаете вы. Ах, как вкусён квес!",
						update.Message.From.ID, bot,
					)
					return err
				}
			default:
				_, err = replyToMsg(update.Message.MessageID, fmt.Sprintf("Что такое %s?", args[1]), update.Message.From.ID, bot)
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
				_, err := replyToMsg(update.Message.MessageID, "Ошибка: пустой ID титула", update.Message.Chat.ID, bot)
				return err
			}
			strID := strings.Join(args[2:], " ")
			i, err := strconv.ParseInt(strID, 10, 64)
			if err != nil {
				_, err = replyToMsg(update.Message.MessageID, "Ошибка: неправильный синтаксис. Синтаксис команды: `о титуле {ID титула}`", update.Message.Chat.ID, bot)
				return err
			} else {
			}
			ID := uint16(i)
			rCount, err := titlesC.CountDocuments(ctx, bson.M{"_id": ID})
			if err != nil {
				return err
			}
			if rCount == 0 {
				_, err = replyToMsg(update.Message.MessageID, fmt.Sprintf("Ошибка: не найдено титула по ID %d", ID), update.Message.Chat.ID, bot)
				return err
			}
			elem := Title{}
			err = titlesC.FindOne(ctx, bson.M{"_id": ID}).Decode(&elem)
			if err != nil {
				return err
			}
			_, err = replyToMsg(update.Message.MessageID, fmt.Sprintf("%s | ID: %d\n%s", elem.Name, ID, elem.Desc), update.Message.Chat.ID, bot)
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
				_, err = replyToMsg(update.Message.MessageID, "У тебя нет вомбата, иди спи сам", update.Message.Chat.ID, bot)
				return err
			} else if womb.Sleep {
				_, err = replyToMsg(update.Message.MessageID, "Твой вомбат уже спит. Если хочешь проснуться, то напиши `проснуться` (логика)", update.Message.Chat.ID, bot)
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
			_, err = replyWithPhoto(update.Message.MessageID, randImg(sleep), "Вы легли спать. Спокойного сна!", update.Message.Chat.ID, bot)
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
				_, err = replyToMsg(update.Message.MessageID, "У тебя нет вомбата, буди себя сам", update.Message.From.ID, bot)
				return err
			} else if !womb.Sleep {
				_, err = replyToMsg(update.Message.MessageID, "Твой вомбат и так не спит, может ты хотел лечь спать? (команда `лечь спать` (опять логика))",
					update.Message.From.ID, bot,
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
					i := uint64(rand.Intn(100) + 1)
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
			_, err = replyWithPhoto(update.Message.MessageID, randImg(unsleep), msg, update.Message.From.ID, bot)
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
				_, err := replyToMsg(update.Message.MessageID, "так и запишем", update.Message.Chat.ID, bot)
				return err
			}
			cargs := args[2:]
			if len(cargs) < 2 {
				_, err := replyToMsg(update.Message.MessageID,
					"Ошибка: вы пропустили аргумент(ы). Синтаксис команды: `перевести шиши [кол-во] [никнейм получателя]`",
					update.Message.Chat.ID, bot,
				)
				return err
			} else if len(cargs) > 2 {
				_, err := replyToMsg(update.Message.MessageID,
					"Ошибка: слишком много аргументов. Синтаксис команды: `перевести шиши [кол-во] [никнейм получателя]`",
					update.Message.Chat.ID, bot,
				)
				return err
			}
			var (
				amount uint64
				err    error
			)
			if amount, err = strconv.ParseUint(cargs[0], 10, 64); err != nil {
				_, err = replyToMsg(
					update.Message.MessageID,
					"нелегальные у Вас какие-то числа",
					update.Message.Chat.ID, bot,
				)
				return err
			}
			var ID int64
			name := cargs[1]
			if len([]rune(name)) > 64 {
				_, err := replyToMsg(update.Message.MessageID, "Слишком длинный никнейм", update.Message.Chat.ID, bot)
				return err
			} else if !isValidName(name) {
				_, err := replyToMsg(update.Message.MessageID, "Нелегальное имя", update.Message.Chat.ID, bot)
				return err
			} else if rCount, err := users.CountDocuments(
				ctx, bson.M{"name": cins(name)}); err != nil {
				return err
			} else if rCount == 0 {
				_, err = replyToMsg(update.Message.MessageID, fmt.Sprintf("Ошибка: вомбата с именем %s не найдено", name), update.Message.Chat.ID, bot)
				return err
			}
			var tWomb User
			err = users.FindOne(ctx, bson.M{"name": cins(name)}).Decode(&tWomb)
			if err != nil {
				return err
			}
			ID = tWomb.ID
			if womb.Money < amount {
				if _, err = strconv.ParseInt(cargs[0], 10, 64); err == nil {
					_, err = replyToMsg(
						update.Message.MessageID, "Ошибка: количество переводимых шишей должно быть больше нуля",
						update.Message.Chat.ID, bot,
					)
					return err
				} else {
					_, err = replyToMsg(update.Message.MessageID, "Ошибка: кол-во переводимых шишей быть числом", update.Message.Chat.ID, bot)
				}
			}
			if amount == 0 {
				_, err = replyToMsg(update.Message.MessageID,
					"Ошибка: количество переводимых шишей должно быть больше нуля",
					update.Message.Chat.ID, bot,
				)
				return err
			}
			if ID == update.Message.From.ID {
				_, err = replyToMsg(update.Message.MessageID, "Ты читер блин нафиг!!!!!! нидам тебе самому себе перевести", update.Message.Chat.ID, bot)
				return err
			}
			rCount, err := users.CountDocuments(ctx, bson.M{"_id": ID})
			if err != nil {
				return err
			}
			if rCount == 0 {
				_, err = replyToMsg(update.Message.MessageID,
					fmt.Sprintf("Ошибка: пользователя с ID %d не найдено", ID),
					update.Message.Chat.ID, bot,
				)
				return err
			}
			womb.Money -= amount
			tWomb.Money += amount
			err = docUpd(tWomb, bson.M{"_id": ID}, users)
			if err != nil {
				return err
			}
			err = docUpd(womb, wombFilter(womb), users)
			if err != nil {
				return err
			}
			_, err = replyToMsg(update.Message.MessageID,
				fmt.Sprintf("Вы успешно перевели %d шишей на счёт %s. Теперь у вас %d шишей при себе",
					amount, tWomb.Name, womb.Money), update.Message.Chat.ID, bot,
			)
			if err != nil {
				return err
			}
			_, err = sendMsg(fmt.Sprintf("Пользователь %s перевёл вам %d шишей. Теперь у вас %d шишей при себе",
				womb.Name, amount, tWomb.Money), ID, bot,
			)
			return err
		},
	},
	{
		Name: "rating",
		Is: func(args []string, update tg.Update) bool {
			return isPrefixInList(strings.ToLower(strings.Join(args, " ")), []string{"рейтинг", "топ"}) && args[0] != "рейтинг" && args[0] != "топ"
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
					_, err := replyToMsg(update.Message.MessageID, fmt.Sprintf("не понимаю, что значит %s", args[1]), update.Message.Chat.ID, bot)
					return err
				}
				if len(args) == 3 {
					if isInList(args[2], []string{"+", "плюс", "++", "увеличение"}) {
						queue = 1
					} else if isInList(args[2], []string{"-", "минус", "--", "уменьшение"}) {
						queue = -1
					} else {
						_, err := replyToMsg(update.Message.MessageID, fmt.Sprintf("не понимаю, что значит %s", args[2]), update.Message.Chat.ID, bot)
						return err
					}
				}
			} else if len(args) != 1 {
				_, err := replyToMsg(update.Message.MessageID, "слишком много аргументов", update.Message.Chat.ID, bot)
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
					msg += fmt.Sprintf("%d | %s | %d шишей при себе\n", num+1, w.Name, w.Money)
				case "xp":
					msg += fmt.Sprintf("%d | %s | %d XP\n", num+1, w.Name, w.XP)
				case "health":
					msg += fmt.Sprintf("%d | %s | %d здоровья\n", num+1, w.Name, w.Health)
				case "force":
					msg += fmt.Sprintf("%d | %s | %d мощи\n", num+1, w.Name, w.Force)
				}
			}
			msg = strings.TrimSuffix(msg, "\n")
			_, err = replyToMsg(update.Message.MessageID, msg, update.Message.Chat.ID, bot)
			return err
		},
	},
}
