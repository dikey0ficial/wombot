package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	// "sort"
	tg "github.com/go-telegram-bot-api/telegram-bot-api"
	jsoniter "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	"strings"
	"time"
)

// Config нужен для настроек
type Config struct {
	Token    string `json:"tg_token"`
	MongoURL string `json:"mongo_url"`
}

// Title — описание титула
type Title struct {
	Name string `bson:"name"`
	Desc string `bson:"desc,omitempty"`
}

// User — описание пользователя
type User struct { // параметры юзера
	ID     int64    `bson:"_id"`
	Name   string   `bson:"name,omitempty"`
	XP     uint32   `bson:"xp"`
	Health uint32   `bson:"health"`
	Force  uint32   `bson:"force"`
	Money  uint64   `bson:"money"`
	Titles []uint16 `bson:"titles"`
	Subs   map[string]int64
}

// Attack реализует атаку
type Attack struct {
	ID   int   `bson:"_id"`
	From int64 `bson:"from"`
	To   int64 `bson:"to"`
}

var ctx = context.TODO()
var json = jsoniter.ConfigCompatibleWithStandardLibrary

// checkerr реализует проверку ошибок без паники
func checkerr(err error) {
	if err != nil && err.Error() != "EOF" {
		log.Panic("ERROR\n\n", err)
	}
}

// checkPanErr реализует проверку ошибку с паникой
func checkPanErr(err error) {
	if err != nil && err.Error() != "EOF" {
		panic(err)
	}
}

// loadConfig нужуен для загрузки конфига до инициализвации требующих его функций
func loadConfig() Config {
	file, err := os.Open("config.json")
	if err != nil && err.Error() != "EOF" {
		checkerr(err)
		return Config{}
	}
	defer file.Close()
	var result = Config{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&result)
	checkPanErr(err)
	return result
}

var conf = loadConfig()

// isInList нужен для проверки сообщений
func isInList(str string, list []string) bool {
	for _, elem := range list {
		if strings.ToLower(str) == strings.ToLower(elem) {
			return true
		}
	}
	return false
}

// isInSubs _
func isInSubs(sub int64, arr map[string]int64) (bool, string) {
	for alias, elem := range arr {
		if sub == elem {
			return true, alias
		}
	}
	return false, ""
}

// hasTitle _
func hasTitle(i uint16, list []uint16) bool {
	for _, elem := range list {
		if i == elem {
			log.Println(elem)
			return true
		}
	}
	return false
}

// toDoc _
func toDoc(v interface{}) (doc *bson.D, err error) {
	data, err := bson.Marshal(v)
	if err != nil {
		return
	}

	err = bson.Unmarshal(data, &doc)
	return
}

// docUpd _
func docUpd(v User, filter bson.D, col mongo.Collection) {
	doc, err := toDoc(v)
	checkerr(err)
	ctx := context.TODO()
	_, err = col.UpdateOne(ctx, filter, bson.M{"$set": doc})
}

// sendMsg отправляет обычное сообщение
func sendMsg(message string, chatID int64, bot *tg.BotAPI) int {
	msg := tg.NewMessage(chatID, message)
	mess, err := bot.Send(msg)
	checkerr(err)
	return mess.MessageID
}

// sendMsgMD отправляет сообщение с markdown
func sendMsgMD(message string, chatID int64, bot *tg.BotAPI) int {
	msg := tg.NewMessage(chatID, message)
	mess, err := bot.Send(msg)
	msg.ParseMode = "markdown"
	checkerr(err)
	return mess.MessageID
}

// replyToMsg отвечает обычным сообщением
func replyToMsg(replyID int, message string, chatID int64, bot *tg.BotAPI) int {
	msg := tg.NewMessage(chatID, message)
	msg.ReplyToMessageID = replyID
	mess, err := bot.Send(msg)
	checkerr(err)
	return mess.MessageID
}

// replyToMsgMD отвечает сообщением с markdown
func replyToMsgMD(replyID int, message string, chatID int64, bot *tg.BotAPI) int {
	msg := tg.NewMessage(chatID, message)
	msg.ReplyToMessageID = replyID
	msg.ParseMode = "markdown"
	mess, err := bot.Send(msg)
	checkerr(err)
	return mess.MessageID
}

// delMsg удаляет сообщение
func delMsg(ID int, chatID int64, bot *tg.BotAPI) {
	deleteMessageConfig := tg.DeleteMessageConfig{
		ChatID:    chatID,
		MessageID: ID,
	}
	_, err := bot.DeleteMessage(deleteMessageConfig)
	checkerr(err)
}

// isInAttacks возвращает 1) есть ли он в атаках 2) являлся ли он from
func isInAttacks(id int64, a *mongo.Collection) (bool, bool) {
	cFrom, err := a.CountDocuments(ctx, bson.D{{"from", id}})
	checkerr(err)
	isFrom := cFrom != 0
	cTo, err := a.CountDocuments(ctx, bson.D{{"to", id}})
	checkerr(err)
	isTo := cTo != 0
	if isFrom || isTo {
		return true, isFrom
	}
	return false, false
}

var standartNicknames []string = []string{"Вомбатыч", "Вомбатус", "wombatkiller2007", "wombatik", "батвом", "Табмов", "Вомбабушка", "womboba"}

func main() {
	mongoClient, err := mongo.NewClient(options.Client().ApplyURI(conf.MongoURL))
	checkPanErr(err)
	err = mongoClient.Connect(ctx)
	checkPanErr(err)
	defer mongoClient.Disconnect(ctx)

	db := *(mongoClient.Database("wombot"))

	users := *(db.Collection("users"))

	titles := []Title{}

	titlesC := *(db.Collection("titles"))
	cur, err := titlesC.Find(ctx, bson.D{})
	defer cur.Close(ctx)
	checkerr(err)
	for cur.Next(ctx) {
		var nextOne Title
		err := cur.Decode(&nextOne)
		checkPanErr(err)
		titles = append(titles, nextOne)
	}
	cur.Close(ctx)

	bot, err := tg.NewBotAPI(conf.Token)
	checkPanErr(err)

	u := tg.NewUpdate(0)
	u.Timeout = 1
	updates, err := bot.GetUpdatesChan(u)
	checkPanErr(err)
	log.Println("Start!")

	for update := range updates {
		if update.Message == nil {
			continue
		}
		if update.Message.Chat.ID != int64(update.Message.From.ID) {
			go func(update tg.Update, titles []Title, titlesC mongo.Collection, bot *tg.BotAPI) {
				peer, from := update.Message.Chat.ID, update.Message.From.ID
				txt, messID := update.Message.Text, update.Message.MessageID
				users = *(db.Collection("users"))

				womb := User{}
				wFil := bson.D{{"_id", from}}
				rCount, err := users.CountDocuments(ctx, wFil)
				checkerr(err)
				isInUsers := rCount != 0
				if isInUsers {
					err = users.FindOne(ctx, wFil).Decode(&womb)
					checkerr(err)
				}

				log.Println("группа", peer, from, update.Message.From.UserName, womb.Name, txt)
				if strings.HasPrefix(strings.ToLower(txt), "о вомбате") {
					strID := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(txt), "о вомбате"))
					var (
						ID    int64
						tWomb User
						ok    bool = true
					)
					if strID == "" {
						if isInUsers {
							ID = int64(from)
							tWomb = womb
						} else {
							replyToMsg(messID, "У вас нет вомбата", peer, bot)
							return
						}
					} else if ID, err = strconv.ParseInt(strID, 10, 64); err == nil {
						rCount, err = users.CountDocuments(ctx, bson.D{{"_id", ID}})
						checkerr(err)
						if rCount == 0 {
							replyToMsg(messID, fmt.Sprintf("Ошибка: пользователя с ID %d не существует", ID), peer, bot)
							return
						}
						err = users.FindOne(ctx, bson.D{{"_id", ID}}).Decode(&tWomb)
						checkerr(err)
					} else if ID, ok = womb.Subs[strID]; ok {
						err = users.FindOne(ctx, bson.D{{"_id", womb.Subs[strID]}}).Decode(&tWomb)
						checkerr(err)
					} else if !ok {
						replyToMsg(messID, fmt.Sprintf("Ошибка: подписчика с алиасом `%s` не найдено", strID), peer, bot)
						return
					} else {
						replyToMsg(messID, "Ошибка: непредвиденная ситуация. Перешлите это сообщение @dikey0ficial\n\nabout womb: else", peer, bot)
						return
					}
					strTitles := ""
					tCount := len(tWomb.Titles)
					if tCount != 0 {
						for _, id := range tWomb.Titles {
							elem := Title{}
							err = titlesC.FindOne(ctx, bson.D{{"_id", id}}).Decode(&elem)
							checkerr(err)
							strTitles += fmt.Sprintf("%s (ID: %d) | ", elem.Name, id)
						}
						strTitles = strings.TrimSuffix(strTitles, " | ")
					} else {
						strTitles = "нет"
					}
					link := fmt.Sprintf("tg://user?id=%d", ID)
					replyToMsgMD(messID, fmt.Sprintf("Вомбат  [%s](%s)  (ID: %d)\nТитулы: %s\n 🕳 %d XP \n ❤ %d здоровья \n ⚡ %d мощи \n 💰 %d шишей", tWomb.Name, link, ID, strTitles, tWomb.XP, tWomb.Health, tWomb.Force, tWomb.Money), peer, bot)
				} else if strings.HasPrefix(strings.ToLower(txt), "хрю") {
					mID := replyToMsg(messID, "АХТУНГ ШВАЙНЕ ШВАЙНЕ ШВАЙНЕ ШВАЙНЕ ААААААА", peer, bot)
					time.Sleep(2 * time.Second)
					delMsg(mID, peer, bot)
				} else if isInList(txt, []string{"помощь", "хелп", "help", "команды", "/help", "/help@wombatobot"}) {
					replyToMsg(messID, "https://telegra.ph/Pomoshch-10-28", peer, bot)
				} else if isInList(txt, []string{"старт", "начать", "/старт", "/start", "/start@wombatobot", "start", "привет"}) {
					replyToMsg(messID, "В групповые чаты писать вомботу НЕ НАДО, он сделан для лс! Пишите в лс: @wombatobot", peer, bot)
				} else if strings.HasPrefix(strings.ToLower(txt), "о титуле") {
					strID := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(txt), "о титуле"))
					if strID == "" {
						replyToMsg(messID, "Ошибка: пустой ID титула", peer, bot)
					} else if i, err := strconv.ParseInt(strID, 10, 64); err == nil {
						checkerr(err)
						ID := uint16(i)
						rCount, err := titlesC.CountDocuments(ctx, bson.D{{"_id", ID}})
						checkerr(err)
						if rCount != 0 {
							elem := Title{}
							err = titlesC.FindOne(ctx, bson.D{{"_id", ID}}).Decode(&elem)
							replyToMsg(messID, fmt.Sprintf("%s | ID: %d\n%s", elem.Name, ID, elem.Desc), peer, bot)
						} else {
							replyToMsg(messID, fmt.Sprintf("Ошибка: не найдено титула по ID %d", ID), peer, bot)
						}
					} else {
						sendMsg("Ошибка: неправильный синтаксис. Синтаксис команды: `о титуле {ID титула}`", peer, bot)
					}
				} else if strings.HasPrefix(strings.ToLower(txt), "о вомботе") {
					replyToMsgMD(messID, "https://telegra.ph/O-vombote-10-29\n**если вы хотели узнать характеристики вомбата, используйте команду `о вомбате`**", peer, bot)
				}
			}(update, titles, titlesC, bot)
			continue
		}
		go func(update tg.Update, titles []Title, titlesC mongo.Collection, bot *tg.BotAPI) {
			peer, from := update.Message.Chat.ID, update.Message.From.ID
			txt, messID := update.Message.Text, update.Message.MessageID
			if from == messID {
				log.Println("AAAAAAAAAAAAAA")
			}
			users = *(db.Collection("users"))

			womb := User{}

			wFil := bson.D{{"_id", peer}}

			rCount, err := users.CountDocuments(ctx, wFil)
			checkerr(err)
			isInUsers := rCount != 0
			if isInUsers {
				err = users.FindOne(ctx, wFil).Decode(&womb)
				checkerr(err)
			}

			log.Println(peer, update.Message.From.UserName, womb.Name, txt)

			if isInList(txt, []string{"старт", "начать", "/старт", "/start", "/start@wombatobot", "start", "привет"}) {
				if isInUsers {
					sendMsg(fmt.Sprintf("Здравствуйте, %s!", womb.Name), peer, bot)
				} else {
					sendMsg("Привет! \n — Завести вомбата: `взять вомбата`\n — Помощь: https://telegra.ph/Pomoshch-10-28 (/help)\n — Канал бота, где есть нужная инфа: @wombatobot_channel\n Приятной игры!", peer, bot)
				}
			} else if isInList(txt, []string{"взять вомбата", "купить вомбата у арабов", "хочу вомбата"}) {
				if isInUsers {
					sendMsg("У тебя как бы уже есть вомбат лолкек. Если хочешь от него избавиться, то напиши `приготовить шашлык`", peer, bot)
				} else {
					rand.Seed(peer)
					newWomb := User{
						ID:     peer,
						Name:   standartNicknames[rand.Intn(len(standartNicknames))],
						XP:     0,
						Health: 5,
						Force:  2,
						Money:  10,
						Titles: []uint16{},
						Subs:   map[string]int64{},
					}
					_, err = users.InsertOne(ctx, &newWomb)
					checkerr(err)

					sendMsg(fmt.Sprintf("Поздравляю, у тебя появился вомбат! Ему выдалось имя `%s`. Ты можешь поменять имя командой `Поменять имя [имя]` за 3 монеты", newWomb.Name), peer, bot)
				}
			} else if strings.HasPrefix(strings.ToLower(txt), "devtools") {
				if hasTitle(0, womb.Titles) {
					cmd := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(txt), "devtools"))
					if strings.HasPrefix(cmd, "set money") {
						strNewMoney := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(cmd), "set money"))
						if i, err := strconv.ParseUint(strNewMoney, 10, 64); err == nil {
							checkerr(err)
							womb.Money = i
							docUpd(womb, wFil, users)
							sendMsg(fmt.Sprintf("Операция проведена успешно! Шишей на счету: %d", womb.Money), peer, bot)
						} else {
							sendMsg("Ошибка: неправильный синтаксис. Синтаксис команды: `devtools set money {кол-во шишей}`", peer, bot)
						}
					} else if strings.HasPrefix(cmd, "reset") {
						arg := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(cmd), "reset"))
						switch arg {
						case "force":
							womb.Force = 2
							docUpd(womb, wFil, users)
							sendMsg("Операция произведена успешно!", peer, bot)
						case "health":
							womb.Health = 5
							docUpd(womb, wFil, users)
							sendMsg("Операция произведена успешно!", peer, bot)
						case "xp":
							womb.XP = 0
							docUpd(womb, wFil, users)
							sendMsg("Операция произведена успешно!", peer, bot)
						case "all":
							womb.Force = 2
							womb.Health = 5
							womb.XP = 0
							docUpd(womb, wFil, users)
							sendMsg("Операция произведена успешно!", peer, bot)
						default:
							sendMsg("Ошибка: неправильный синтаксис. Синтаксис команды: `devtools reset [force/health/xp/all]`", peer, bot)
						}
					} else if cmd == "help" {
						sendMsg("https://telegra.ph/Vombot-devtools-help-10-28", peer, bot)
					}
				} else if strings.TrimSpace(txt) == "devtools on" {
					womb.Titles = append(womb.Titles, 0)
					docUpd(womb, wFil, users)
					sendMsg("Выдан титул \"Вомботестер\" (ID: 0)", peer, bot)
				}
			} else if isInList(txt, []string{"приготовить шашлык", "продать вомбата арабам", "слить вомбата в унитаз"}) {
				if isInUsers {
					if !(hasTitle(1, womb.Titles)) {
						_, err = users.DeleteOne(ctx, wFil)
						checkerr(err)
						sendMsg("Вы уничтожили вомбата в количестве 1 штука. Вы - нехорошее существо", peer, bot)
					} else {
						sendMsg("Ошибка: вы лишены права уничтожать вомбата; обратитксь к @dikey0ficial за разрешением", peer, bot)
					}
				} else {
					sendMsg("Но у вас нет вомбата...", peer, bot)
				}
			} else if strings.HasPrefix(strings.ToLower(txt), "поменять имя") {
				if isInUsers {
					name := strings.Title(strings.TrimSpace(strings.TrimPrefix(strings.ToLower(txt), "поменять имя ")))
					if womb.Money >= 3 {
						if isInList(name, []string{"admin", "вoмбoт", "вoмбoт", "вомбoт", "вомбот", "бот", "bot", "бoт", "bоt"}) {
							sendMsg("Такие никнеймы заводить нельзя", peer, bot)
						} else if name != "" {
							womb.Money -= 3
							split := strings.Fields(txt)
							caseName := strings.Join(split[2:], " ")
							womb.Name = caseName
							docUpd(womb, wFil, users)

							sendMsg(fmt.Sprintf("Теперь вашего вомбата зовут %s. С вашего счёта сняли 3 шиша", caseName), peer, bot)
						} else {
							sendMsg("У вас пустое имя...", peer, bot)
						}
					} else {
						sendMsg("Мало шишей блин нафиг!!!!", peer, bot)
					}
				} else {
					sendMsg("Да блин нафиг, вы вобмата забыли завести!!!!!!!", peer, bot)
				}
			} else if isInList(txt, []string{"помощь", "хелп", "help", "команды", "/help", "/help@wombatobot"}) {
				sendMsg("https://telegra.ph/Pomoshch-10-28", peer, bot)
			} else if isInList(txt, []string{"купить здоровье", "прокачка здоровья", "прокачать здоровье"}) {
				if isInUsers {
					if womb.Money >= 5 {
						if uint64(womb.Health+1) < 2^32 {
							womb.Money -= 5
							womb.Health++
							docUpd(womb, wFil, users)
							sendMsg(fmt.Sprintf("Поздравляю! Теперь у вас %d здоровья и %d шишей", womb.Health, womb.Money), peer, bot)
						} else {
							sendMsg("Ошибка: вы достигли максимального количества здоровья (2 в 32 степени). Если это вас возмущает, обратитесь к @dikey0ficial", peer, bot)
						}
					} else {
						sendMsg("Надо накопить побольше шишей! 1 здоровье = 5 шишей", peer, bot)
					}
				} else {
					sendMsg("У тя ваще вобата нет...", peer, bot)
				}
			} else if isInList(txt, []string{"купить мощь", "прокачка мощи", "прокачка силы", "прокачать мощь", "прокачать силу"}) {
				if isInUsers {
					if womb.Money >= 3 {
						if uint64(womb.Force+1) < 2^32 {
							womb.Money -= 3
							womb.Force++
							docUpd(womb, wFil, users)
							sendMsg(fmt.Sprintf("Поздравляю! Теперь у вас %d мощи и %d шишей", womb.Force, womb.Money), peer, bot)
						} else {
							sendMsg("Ошибка: вы достигли максимального количества здоровья (2 в 32 степени). Если это вас возмущает, обратитесь к @dikey0ficial", peer, bot)
						}
					} else {
						sendMsg("Надо накопить побольше шишей! 1 мощь = 3 шиша", peer, bot)
					}
				} else {
					sendMsg("У тя ваще вобата нет...", peer, bot)
				}
			} else if isInList(txt, []string{"поиск денег"}) {
				if isInUsers {
					if womb.Money >= 1 {
						womb.Money--
						rand.Seed(time.Now().UnixNano())
						if ch := rand.Int(); ch%2 == 0 || hasTitle(2, womb.Titles) && (ch%2 == 0 || ch%3 == 0) {
							rand.Seed(time.Now().UnixNano())
							win := rand.Intn(9) + 1
							womb.Money += uint64(win)
							if addXP := rand.Intn(512 - 1); addXP < 5 {
								womb.XP += uint32(addXP)
								sendMsg(fmt.Sprintf("Поздравляем! Вы нашли на дороге %d шишей, а ещё вам дали %d XP! Теперь у вас %d шишей и %d XP", win, addXP, womb.Money, womb.XP), peer, bot)
							} else {
								sendMsg(fmt.Sprintf("Поздравляем! Вы нашли на дороге %d шишей! Теперь их у вас %d", win, womb.Money), peer, bot)
							}
						} else {
							sendMsg("Вы заплатили один шиш охранникам денежной дорожки, но увы, вы так ничего и не нашли", peer, bot)
						}
						docUpd(womb, wFil, users)

					} else {
						sendMsg("Охранники тебя прогнали; они требуют шиш за проход, а у тебя и шиша-то нет", peer, bot)
					}
				} else {
					sendMsg("А ты куда? У тебя вомбата нет...", peer, bot)
				}
			} else if strings.HasPrefix(strings.ToLower(txt), "о титуле") {
				strID := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(txt), "о титуле"))
				if strID == "" {
					sendMsg("Ошибка: пустой ID титула", peer, bot)
				} else if i, err := strconv.ParseInt(strID, 10, 64); err == nil {
					checkerr(err)
					ID := uint16(i)
					rCount, err := titlesC.CountDocuments(ctx, bson.D{{"_id", ID}})
					checkerr(err)
					if rCount != 0 {
						elem := Title{}
						err = titlesC.FindOne(ctx, bson.D{{"_id", ID}}).Decode(&elem)
						sendMsg(fmt.Sprintf("%s | ID: %d\n%s", elem.Name, ID, elem.Desc), peer, bot)
					} else {
						sendMsg(fmt.Sprintf("Ошибка: не найдено титула по ID %d", ID), peer, bot)
					}
				} else {
					sendMsg("Ошибка: неправильный синтаксис. Синтаксис команды: `о титуле {ID титула}`", peer, bot)
				}
			} else if strings.HasPrefix(strings.ToLower(txt), "подписаться") {
				args := strings.Fields(strings.TrimSpace(strings.TrimPrefix(strings.ToLower(txt), "подписаться")))
				if len(args) == 0 {
					sendMsg("Ошибка: пропущены аргументы `ID` и `алиас`. Синтаксис команды: `подписаться [ID] [алиас (без пробелов)]`", peer, bot)
				} else if len(args) == 1 {
					sendMsg("Ошибка: пропущен аргумент `алиас`. Синтаксис команды: `подписаться [ID] [алиас (без пробелов)]`", peer, bot)
				} else if len(args) == 2 {
					if ID, err := strconv.ParseInt(args[0], 10, 64); err == nil {
						if _, err := strconv.ParseInt(args[1], 10, 64); err == nil {
							sendMsg("Ошибка: алиас не должен быть числом", peer, bot)
						} else {
							if elem, ok := womb.Subs[args[1]]; !ok {
								rCount, err = users.CountDocuments(ctx, bson.D{{"_id", ID}})
								checkerr(err)
								subbed, name := isInSubs(ID, womb.Subs)
								if subbed {
									sendMsg(fmt.Sprintf("Ошибка: вы уже подписались на вомбата с ID %d (алиас: %s). Для того, чтобы отписаться, напишите команду \"отписаться %s\"", ID, name, name), peer, bot)
									return
								}
								if rCount != 0 {
									womb.Subs[args[1]] = ID
									docUpd(womb, wFil, users)
									sendMsg(fmt.Sprintf("Вомбат с ID %d добавлен в ваши подписки", ID), peer, bot)
								} else {
									sendMsg(fmt.Sprintf("Ошибка: пользователя с ID %d не найдено", ID), peer, bot)
								}
							} else {
								sendMsg(fmt.Sprintf("Ошибка: алиас %s занят id %d", args[1], elem), peer, bot)
							}
						}
					} else {
						sendMsg(fmt.Sprintf("Ошибка: `%s` не является целым числом", args[0]), peer, bot)
					}
				} else {
					sendMsg("Ошибка: слишком много аргументов. Синтаксис команды: `подписаться [ID] [алиас (без пробелов)]", peer, bot)
				}
			} else if strings.HasPrefix(strings.ToLower(txt), "отписаться") {
				alias := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(txt), "отписаться"))
				if _, ok := womb.Subs[alias]; ok {
					delete(womb.Subs, alias)
					docUpd(womb, wFil, users)

					sendMsg(fmt.Sprintf("Вы отписались от пользователя с алиасом %s", alias), peer, bot)
				} else {
					sendMsg(fmt.Sprintf("Ошибка: вы не подписаны на пользователя с алиасом `%s`", alias), peer, bot)
				}
			} else if isInList(txt, []string{"подписки", "мои подписки", "список подписок"}) {
				strSubs := "Вот список твоих подписок:"
				if len(womb.Subs) != 0 {
					for alias, id := range womb.Subs {
						rCount, err = users.CountDocuments(ctx, bson.D{{"_id", id}})
						checkerr(err)
						if rCount != 0 {
							tWomb := User{}
							err = users.FindOne(ctx, bson.D{{"_id", id}}).Decode(&tWomb)
							checkerr(err)
							strSubs += fmt.Sprintf("\n %s | ID: %d | Алиас: %s", tWomb.Name, id, alias)
						} else {
							strSubs += fmt.Sprintf("\n Ошибка: пользователь по алиасу `%s` не найден", alias)
						}
					}
				} else {
					strSubs = "У тебя пока ещё нет подписок"
				}
				sendMsg(strSubs, peer, bot)
			} else if isInList(txt, []string{"мои вомбаты", "мои вомбатроны", "вомбатроны", "лента подписок"}) {
				if len(womb.Subs) == 0 {
					sendMsg("У тебя пока ещё нет подписок", peer, bot)
					return
				}
				for alias, ID := range womb.Subs {
					rCount, err := users.CountDocuments(ctx, bson.D{{"_id", ID}})
					checkerr(err)
					if rCount != 0 {
						tWomb := User{}
						err = users.FindOne(ctx, bson.D{{"_id", ID}}).Decode(&tWomb)
						checkerr(err)
						strTitles := ""
						tCount := len(tWomb.Titles)
						if tCount != 0 {
							for _, id := range tWomb.Titles {
								elem := Title{}
								err = titlesC.FindOne(ctx, bson.D{{"_id", id}}).Decode(&elem)
								checkerr(err)
								strTitles += fmt.Sprintf("%s (ID: %d) | ", elem.Name, id)
							}
							strTitles = strings.TrimSuffix(strTitles, " | ")
						} else {
							strTitles = "нет"
						}
						sendMsg(fmt.Sprintf("Вомбат  %s (ID: %d; Алиас: %s)\nТитулы: %s\n 🕳 %d XP \n ❤ %d здоровья \n ⚡ %d мощи \n 💰 %d шишей", tWomb.Name, ID, alias, strTitles, tWomb.XP, tWomb.Health, tWomb.Force, tWomb.Money), peer, bot)
					} else {
						sendMsg(fmt.Sprintf("Ошибка: подписчика с алиасом `%s` не обнаружено", alias), peer, bot)
					}
				}
			} else if strings.HasPrefix(strings.ToLower(txt), "о вомбате") {
				strID := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(txt), "о вомбате"))
				var (
					ID    int64
					tWomb User
					ok    bool
				)
				if strID == "" {
					if isInUsers {
						ID = peer
						tWomb = womb
					} else {
						sendMsg("У вас нет вомбата", peer, bot)
						return
					}
				} else if ID, err = strconv.ParseInt(strID, 10, 64); err == nil {
					rCount, err = users.CountDocuments(ctx, bson.D{{"_id", ID}})
					checkerr(err)
					if rCount == 0 {
						sendMsg(fmt.Sprintf("Ошибка: пользователя с ID %d не существует", ID), peer, bot)
						return
					}
					err = users.FindOne(ctx, bson.D{{"_id", ID}}).Decode(&tWomb)
					checkerr(err)
				} else if ID, ok = womb.Subs[strID]; ok {
					err = users.FindOne(ctx, bson.D{{"_id", womb.Subs[strID]}}).Decode(&tWomb)
					checkerr(err)
				} else if !ok {
					replyToMsg(messID, fmt.Sprintf("Ошибка: подписчика с алиасом `%s` не найдено", strID), peer, bot)
					return
				} else {
					replyToMsg(messID, "Ошибка: непредвиденная ситуация. Перешлите это сообщение @dikey0ficial\n\nabout womb: else", peer, bot)
					return
				}
				strTitles := ""
				tCount := len(tWomb.Titles)
				if tCount != 0 {
					for _, id := range tWomb.Titles {
						elem := Title{}
						err = titlesC.FindOne(ctx, bson.D{{"_id", id}}).Decode(&elem)
						checkerr(err)
						strTitles += fmt.Sprintf("%s (ID: %d) | ", elem.Name, id)
					}
					strTitles = strings.TrimSuffix(strTitles, " | ")
				} else {
					strTitles = "нет"
				}
				sendMsgMD(fmt.Sprintf("Вомбат  %s (ID: %d)\nТитулы: %s\n 🕳 %d XP \n ❤ %d здоровья \n ⚡ %d мощи \n 💰 %d шишей", tWomb.Name, ID, strTitles, tWomb.XP, tWomb.Health, tWomb.Force, tWomb.Money), peer, bot)
			} else if strings.HasPrefix(strings.ToLower(txt), "о вомботе") {
				sendMsgMD("https://telegra.ph/O-vombote-10-29\n**если вы хотели узнать характеристики вомбата, используйте команду `о вомбате`**", peer, bot)
			} else if strings.HasPrefix(strings.ToLower(txt), "перевести шиши") {
				args := strings.Fields(strings.TrimSpace(strings.TrimPrefix(strings.ToLower(txt), "перевести шиши")))
				if len(args) < 2 {
					sendMsg("Ошибка: вы пропустили аргумент(ы). Синтаксис команды: `перевести шиши [кол-во] [ID/алиас получателя]`", peer, bot)
				} else if len(args) > 2 {
					sendMsg("Ошибка: слишком много аргументов. Синтаксис команды: `перевести шиши [кол-во] [ID/алиас получателя]`", peer, bot)
				} else {
					if amount, err := strconv.ParseUint(args[0], 10, 64); err == nil {
						var ID int64
						if ID, err = strconv.ParseInt(args[1], 10, 64); err != nil {
							var ok bool
							if ID, ok = womb.Subs[args[1]]; !ok {
								sendMsg(fmt.Sprintf("Ошибка: алиаса %s не обнаружено", args[1]), peer, bot)
								return
							}
						}
						if womb.Money > amount {
							if amount != 0 {
								if ID == peer {
									sendMsg("Ты читер блин нафиг!!!!!! нидам тебе самому себе перевести", peer, bot)
									return
								}
								rCount, err = users.CountDocuments(ctx, bson.D{{"_id", ID}})
								checkerr(err)
								if rCount != 0 {
									tWomb := User{}
									err = users.FindOne(ctx, bson.D{{"_id", ID}}).Decode(&tWomb)
									checkerr(err)
									womb.Money -= amount
									tWomb.Money += amount
									docUpd(tWomb, bson.D{{"_id", ID}}, users)
									docUpd(womb, wFil, users)
									sendMsg(fmt.Sprintf("Вы успешно перевели %d шишей на счёт %s. Теперь у вас %d шишей", amount, tWomb.Name, womb.Money), peer, bot)
									sendMsg(fmt.Sprintf("Пользователь %s (ID: %d) перевёл вам %d шишей. Теперь у вас %d шишей", womb.Name, peer, amount, tWomb.Money), ID, bot)
								} else {
									sendMsg(fmt.Sprintf("Ошибка: пользователя с ID %d не найдено", ID), peer, bot)
								}
							} else {
								sendMsg("Ошибка: количество переводимых шишей должно быть больше нуля", peer, bot)
							}
						} else {
							sendMsg(fmt.Sprintf("Ошибка: размер перевода (%d) должен быть меньше кол-ва ваших шишей (%d)", amount, womb.Money), peer, bot)
						}
					} else {
						if _, err := strconv.ParseInt(args[0], 10, 64); err == nil {
							sendMsg("Ошибка: количество переводимых шишей должно быть больше нуля", peer, bot)
						} else {
							sendMsg("Ошибка: кол-во переводимых шишей быть числом", peer, bot)
						}
					}
				}
			} else if txt == "обновить данные" && hasTitle(0, womb.Titles) {
				users = *(db.Collection("users"))
				titlesC := *(db.Collection("titles"))
				cur, err := titlesC.Find(ctx, bson.D{})
				defer cur.Close(ctx)
				checkerr(err)
				for cur.Next(ctx) {
					var nextOne Title
					err := cur.Decode(&nextOne)
					checkPanErr(err)
					titles = append(titles, nextOne)
				}
				cur.Close(ctx)
				sendMsg("Успешно обновлено!", peer, bot)
			} else if isInList(txt, []string{"купить квес", "купить квесс", "купить qwess", "попить квес", "попить квесс", "попить qwess"}) {
				if isInUsers {
					if womb.Money >= 256 {
						if !(hasTitle(2, womb.Titles)) {
							womb.Titles = append(womb.Titles, 2)
							womb.Money -= 256
							docUpd(womb, wFil, users)
							sendMsg("Вы купили чудесного вкуса квес у кролика-Лепса в ларьке за 256 шишей. Глотнув этот напиток, вы поняли, что получили новый титул с ID 2", peer, bot)
						} else {
							womb.Money -= 256
							docUpd(womb, wFil, users)
							sendMsg("Вы вновь купили вкусного квеса у того же кролика-Лепса в том же ларьке за 256 шишей. \"Он так освежает, я чувствую себя человеком\" — думаете вы. Ах, как вкусён квес!", peer, bot)
						}
					} else {
						sendMsg("Вы подошли к ближайшему ларьку, но, увы, кролик-Лепс на кассе сказал, что надо 256 шишей, а у вас, к сожалению, меньше", peer, bot)
					}
				} else {
					sendMsg("К сожалению, вам нужны шиши, чтобы купить квес, а шиши есть только у вомбатов...", peer, bot)
				}
			}
		}(update, titles, titlesC, bot)
	}
}
