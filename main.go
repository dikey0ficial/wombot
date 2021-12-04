package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	// "sort"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
	Token     string `json:"tg_token"`
	MongoURL  string `json:"mongo_url"`
	SupChatID int64  `json:"support_chat_id"`
	Debug     bool   `json:"debug"`
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
	Sleep  bool `bson:"sleep"`
}

// Attack реализует атаку
type Attack struct {
	ID   string `bson:"_id"`
	From int64  `bson:"from"`
	To   int64  `bson:"to"`
}

// Imgs реализует группу картинок
type Imgs struct {
	ID     string   `bson:"_id"`
	Images []string `bson:"imgs"`
}

// Banked реализет вомбанковскую ячейку
type Banked struct {
	ID    int64  `bson:"_id"`
	Money uint64 `bson:"money"`
}

var ctx = context.TODO()
var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	logpath = "runtime.log"
)

func initLog() *log.Logger {
	f, err := os.OpenFile(logpath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	nlog := log.New(f, "", log.Ldate|log.Ltime)
	go func(f *os.File) {
		// данный костыль нужен, чтобы файл
		// не закрывался раньше времени,
		// но при этом всё же
		// инициализировать логгер
		defer f.Close()
		for {
			time.Sleep(300 * time.Second)
		}
	}(f)
	return nlog
}

var rlog = initLog()

// checkerr реализует проверку ошибок без паники
func checkerr(err error) {
	if err != nil && err.Error() != "EOF" {
		fmt.Printf("error! %v", err)
		rlog.Panicf("ERROR %v", err)
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
	checkerr(err)
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
			rlog.Println(elem)
			return true
		}
	}
	return false
}

func isPrefixInList(txt string, list []string) bool {
	var is bool = false
	for _, val := range list {
		is = strings.HasPrefix(strings.ToLower(txt), val)
		if is {
			break
		}
	}
	return is
}

// toDoc _
func toDoc(v interface{}) (doc *bson.M, err error) {
	data, err := bson.Marshal(v)
	if err != nil {
		return
	}

	err = bson.Unmarshal(data, &doc)
	return
}

// docUpd _
func docUpd(v interface{}, filter bson.M, col *mongo.Collection) error {
	doc, err := toDoc(v)
	if err != nil {
		return err
	}
	ctx := context.TODO()
	_, err = col.UpdateOne(ctx, filter, bson.M{"$set": doc})
	if err != nil {
		return err
	}
	return nil
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

// replyToMsgMDNL отвечает сообщением с markdown без ссылок
func replyToMsgMDNL(replyID int, message string, chatID int64, bot *tg.BotAPI) int {
	msg := tg.NewMessage(chatID, message)
	msg.ReplyToMessageID = replyID
	msg.ParseMode = "markdown"
	msg.DisableWebPagePreview = true
	mess, err := bot.Send(msg)
	checkerr(err)
	return mess.MessageID
}

// sendPhoto отправляет текст с картинкой
func sendPhoto(id, caption string, chatID int64, bot *tg.BotAPI) int {
	msg := tg.NewPhoto(chatID, tg.FileID(id))
	msg.Caption = caption
	mess, err := bot.Send(msg)
	checkerr(err)
	return mess.MessageID
}

// sendPhotoMD отправляет текст с markdown с картинкой
func sendPhotoMD(id, caption string, chatID int64, bot *tg.BotAPI) int {
	msg := tg.NewPhoto(chatID, tg.FileID(id))
	msg.Caption = caption
	msg.ParseMode = "markdown"
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

// replyWithPhotoMD отвечает картинкой с текстом с markdown
func replyWithPhotoMD(replyID int, id, caption string, chatID int64, bot *tg.BotAPI) int {
	msg := tg.NewPhoto(chatID, tg.FileID(id))
	msg.Caption = caption
	msg.ReplyToMessageID = replyID
	msg.ParseMode = "markdown"
	mess, err := bot.Send(msg)
	checkerr(err)
	return mess.MessageID
}

// replyWithPhotoMD отвечает картинкой с текстом
func replyWithPhoto(replyID int, id, caption string, chatID int64, bot *tg.BotAPI) int {
	msg := tg.NewPhoto(chatID, tg.FileID(id))
	msg.Caption = caption
	msg.ReplyToMessageID = replyID
	mess, err := bot.Send(msg)
	checkerr(err)
	return mess.MessageID
}

// isInAttacks возвращает информацию, еслть ли существо в атаках и
// отправитель ли он
func isInAttacks(id int64, attacks *mongo.Collection) (isIn, isFrom bool) {
	if f, err := attacks.CountDocuments(ctx, bson.M{"from": id}); f > 0 && err == nil {
		isFrom = true
	} else if err != nil {
		checkerr(err)
	}
	var isTo bool
	if t, err := attacks.CountDocuments(ctx, bson.M{"to": id}); t > 0 && err == nil {
		isTo = true
	} else if err != nil {
		checkerr(err)
	}
	isIn = isFrom || isTo
	return isIn, isFrom
}

var errNoAttack error = fmt.Errorf("there aren't any attacks")

func getAttackByID(aid string, attacks *mongo.Collection) (at Attack, err error) {
	c, err := attacks.CountDocuments(ctx, bson.M{"_id": aid})
	if err != nil {
		return Attack{}, err
	} else if c < 1 {
		return Attack{}, errNoAttack
	}
	err = attacks.FindOne(ctx, bson.M{"_id": aid}).Decode(&at)
	if err != nil {
		return Attack{}, err
	}
	return at, nil
}

func getAttackByWomb(id int64, isFrom bool, attacks *mongo.Collection) (at Attack, err error) {
	var (
		fil bson.M
	)
	if isFrom {
		fil = bson.M{"from": id}
	} else {
		fil = bson.M{"to": id}
	}
	c, err := attacks.CountDocuments(ctx, fil)
	if err != nil {
		return Attack{}, err
	} else if c < 1 {
		return Attack{}, errNoAttack
	}
	err = attacks.FindOne(ctx, fil).Decode(&at)
	if err != nil {
		return Attack{}, err
	}
	return at, nil
}

// delMsg удаляет сообщение
func delMsg(ID int, chatID int64, bot *tg.BotAPI) {
	delConfig := tg.DeleteMessageConfig{
		ChatID:    chatID,
		MessageID: ID,
	}
	_, err := bot.Request(delConfig)
	checkerr(err)
}

func editMsg(mid int, txt string, chatID int64, bot *tg.BotAPI) {
	editConfig := tg.EditMessageTextConfig{
		BaseEdit: tg.BaseEdit{
			ChatID:    chatID,
			MessageID: mid,
		},
		Text: txt,
	}
	_, err := bot.Request(editConfig)
	checkerr(err)
}

var errNoImgs = fmt.Errorf("getImgs: no groups with this name")

func getImgs(imgsC *mongo.Collection, name string) (imgs Imgs, err error) {
	if rCount, err := imgsC.CountDocuments(ctx, bson.M{"_id": name}); err != nil {
		return imgs, err
	} else if rCount == 0 {
		return imgs, errNoImgs
	}
	err = imgsC.FindOne(ctx, bson.M{"_id": name}).Decode(&imgs)
	return imgs, err
}

func randImg(imgs Imgs) string {
	rand.Seed(time.Now().Unix())
	return imgs.Images[rand.Intn(len(imgs.Images))]
}

func getBanked(bank *mongo.Collection, ID int64) (b Banked, err error) {

	return b, err
}

var standartNicknames []string = []string{"Вомбатыч", "Вомбатус", "wombatkiller2007", "wombatik", "батвом", "Табмов",
	"Вомбабушка", "womboba"}

func main() {
	// init
	mongoClient, err := mongo.NewClient(options.Client().ApplyURI(conf.MongoURL))
	checkerr(err)
	err = mongoClient.Connect(ctx)
	checkerr(err)
	defer mongoClient.Disconnect(ctx)

	db := mongoClient.Database("wombot")

	users := db.Collection("users")

	attacks := db.Collection("attacks")

	bank := db.Collection("bank")

	var titles []Title

	titlesC := db.Collection("titles")
	cur, err := titlesC.Find(ctx, bson.M{})
	defer cur.Close(ctx)
	checkerr(err)
	for cur.Next(ctx) {
		var nextOne Title
		err := cur.Decode(&nextOne)
		checkerr(err)
		titles = append(titles, nextOne)
	}
	cur.Close(ctx)

	imgsC := db.Collection("imgs")

	bot, err := tg.NewBotAPI(conf.Token)
	checkerr(err)

	u := tg.NewUpdate(0)
	u.Timeout = 1
	updates := bot.GetUpdatesChan(u)
	checkerr(err)
	fmt.Println("Start!")

	for update := range updates {
		if update.Message == nil {
			continue
		} else if update.Message.Photo != nil && conf.Debug {
			rlog.Println("img ", (update.Message.Photo)[0].FileID)
			continue
		}
		if update.Message.Chat.ID == conf.SupChatID {
			// MESSAGE_ADMIN_CHAT
			go func(update tg.Update, bot *tg.BotAPI) {
				if update.Message.ReplyToMessage == nil || update.Message.ReplyToMessage.From.ID != bot.Self.ID {
					return
				}
				strMessID := strings.Fields(update.Message.ReplyToMessage.Text)[0]
				omID, err := strconv.ParseInt(strMessID, 10, 64)
				if err != nil {
					return
				}
				strPeer := strings.Fields(update.Message.ReplyToMessage.Text)[1]
				peer, err := strconv.ParseInt(strPeer, 10, 64)
				if err != nil {
					return
				}
				if update.Message.From.UserName != "" {
					replyToMsgMDNL(int(omID), fmt.Sprintf(
						"Ответ от [админа](t.me/%s): \n%s",
						update.Message.From.UserName,
						update.Message.Text,
					), peer, bot)
				} else {
					replyToMsgMD(int(omID), fmt.Sprintf(
						"Ответ от админа (для обращений: %d): \n%s",
						update.Message.From.ID,
						update.Message.Text,
					), peer, bot)
				}

			}(update, bot)
			continue
		}
		if update.Message.Chat.ID != int64(update.Message.From.ID) {
			// MESSAGE_GROUP_CHAT
			go func(update tg.Update, titles []Title, bot *tg.BotAPI, users, titlesC,
				attacks, imgsC, bank *mongo.Collection) {

				const errStart string = "Ошибка... Ответьте командой /admin на это сообщение\ngr: "

				peer, from := update.Message.Chat.ID, update.Message.From.ID
				txt, messID := strings.TrimSpace(update.Message.Text), update.Message.MessageID
				users = db.Collection("users")

				womb := User{}
				wFil := bson.M{"_id": from}
				rCount, err := users.CountDocuments(ctx, wFil)
				if err != nil {
					replyToMsg(messID, errStart+"isInUsers: count_womb", peer, bot)
					rlog.Println("Error: ", err)
					return
				}
				isInUsers := rCount != 0
				if isInUsers {
					err = users.FindOne(ctx, wFil).Decode(&womb)
					if err != nil {
						replyToMsg(messID, errStart+"womb: find_womb", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
				}

				rlog.Printf("MESSGAE_GROUP p:%d f:%d un:%s, wn:%s, t:%s\n", peer, from, update.Message.From.UserName, womb.Name, txt)
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
						rCount, err = users.CountDocuments(ctx, bson.M{"_id": ID})
						if err != nil {
							replyToMsg(messID, errStart+"about_womb: isInUsers", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						if rCount == 0 {
							replyToMsg(messID, fmt.Sprintf("Ошибка: пользователя с ID %d не существует", ID), peer, bot)
							return
						}
						err = users.FindOne(ctx, bson.M{"_id": ID}).Decode(&tWomb)
						if err != nil {
							replyToMsg(messID, errStart+"about_womb: find_womb", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
					} else if ID, ok = womb.Subs[strID]; ok {
						err = users.FindOne(ctx, bson.M{"_id": womb.Subs[strID]}).Decode(&tWomb)
						if err != nil {
							replyToMsg(messID, errStart+"about_womb: alias_no_users", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
					} else if !ok {
						if len([]rune(strID)) > 64 {
							replyToMsg(messID, "Слишком длинный алиас...", peer, bot)
							return
						}
						replyToMsg(messID, fmt.Sprintf("Ошибка: подписчика с алиасом `%s` не найдено", strID), peer, bot)
						return
					} else {
						replyToMsg(messID, errStart+"about_womb: else", peer, bot)
						rlog.Println("Error: about_womb: else")
						return
					}
					strTitles := ""
					tCount := len(tWomb.Titles)
					if tCount != 0 {
						for _, id := range tWomb.Titles {
							rCount, err = titlesC.CountDocuments(ctx, bson.M{"_id": id})
							if err != nil {
								replyToMsg(messID, errStart+"list_subs: count_titles", peer, bot)
								rlog.Println("Error: ", err)
								return
							}
							if rCount == 0 {
								strTitles += fmt.Sprintf("Ошибка: титула с ID %d нет (ответьте командой /admin) |", id)
								continue
							}
							elem := Title{}
							err = titlesC.FindOne(ctx, bson.M{"_id": id}).Decode(&elem)
							if err != nil {
								replyToMsg(messID, errStart+"about_womb: title: find_title", peer, bot)
								rlog.Println("Error: ", err)
								return
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
					link := fmt.Sprintf("tg://user?id=%d", ID)
					abimg, err := getImgs(imgsC, "about")
					if err != nil {
						replyToMsg(messID, errStart+"about_womb: get_imgs", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
					replyWithPhotoMD(messID, randImg(abimg), fmt.Sprintf(
						"Вомбат [%s](%s) (ID: %d) {%s}\nТитулы: %s\n 🕳 %d XP \n ❤ %d здоровья \n ⚡ %d мощи \n 💰 %d шишей при себе",
						tWomb.Name, link, ID, sl, strTitles, tWomb.XP, tWomb.Health, tWomb.Force, tWomb.Money),
						peer, bot,
					)
				} else if strings.HasPrefix(strings.ToLower(txt), "хрю") {
					hru, err := getImgs(imgsC, "schweine")
					if err != nil {
						replyToMsg(messID, errStart+"schweine: get_imgs", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
					mID := replyWithPhoto(messID, randImg(hru), "АХТУНГ ШВАЙНЕ ШВАЙНЕ ШВАЙНЕ ШВАЙНЕ ААААААА", peer, bot)
					time.Sleep(15 * time.Second)
					delMsg(mID, peer, bot)
				} else if isInList(txt, []string{"помощь", "команды", "/help@wombatobot"}) {
					replyToMsg(messID, "https://telegra.ph/Pomoshch-10-28", peer, bot)
				} else if isInList(txt, []string{"/старт", "/start@wombatobot"}) {
					replyToMsg(messID, "В групповые чаты писать вомботу НЕ НАДО, он сделан для лс! Пишите в лс: @wombatobot", peer, bot)
				} else if strings.HasPrefix(strings.ToLower(txt), "о титуле") {
					strID := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(txt), "о титуле"))
					if strID == "" {
						replyToMsg(messID, "Ошибка: пустой ID титула", peer, bot)
					} else if i, err := strconv.ParseInt(strID, 10, 64); err == nil {
						ID := uint16(i)
						rCount, err := titlesC.CountDocuments(ctx, bson.M{"_id": ID})
						if err != nil {
							replyToMsg(messID, errStart+"about_title: count_titles", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						if rCount != 0 {
							elem := Title{}
							err = titlesC.FindOne(ctx, bson.M{"_id": ID}).Decode(&elem)
							replyToMsg(messID, fmt.Sprintf("%s | ID: %d\n%s", elem.Name, ID, elem.Desc), peer, bot)
						} else {
							replyToMsg(messID, fmt.Sprintf("Ошибка: не найдено титула по ID %d", ID), peer, bot)
						}
					} else {
						sendMsg("Ошибка: неправильный синтаксис. Синтаксис команды: `о титуле {ID титула}`", peer, bot)
					}
				} else if strings.HasPrefix(strings.ToLower(txt), "о вомботе") {
					replyToMsgMD(messID,
						"https://telegra.ph/O-vombote-10-29\n**если вы хотели узнать характеристики вомбата, используйте команду `о вомбате`**",
						peer, bot,
					)
				} else if isPrefixInList(txt, []string{"/admin", "/админ", "/admin@wombatobot", "одмен!", "/баг", "/bug", "/bug@wombatobot", "/support", "/support@wombatobot"}) {
					oArgs := strings.Fields(strings.ToLower(txt))
					if len(oArgs) < 2 {
						if update.Message.ReplyToMessage == nil {
							replyToMsg(messID, "Ты чаво... где письмо??", peer, bot)
							return
						}
						r := update.Message.ReplyToMessage
						sendMsg(fmt.Sprintf(
							"%d %d \nписьмо из группы (%d @%s) от %d (@%s isInUsers: %v), отвечающее на: \n%s\n(id:%d fr:%d @%s)",
							messID, peer, peer, update.Message.Chat.UserName,
							from, update.Message.From.UserName,
							isInUsers, r.Text, r.MessageID, r.From.ID, r.From.UserName),
							conf.SupChatID, bot,
						)
						replyToMsg(messID, "Письмо отправлено! Скоро (или нет) придёт ответ", peer, bot)
					} else {
						if update.Message.ReplyToMessage == nil {
							msg := strings.Join(oArgs[1:], " ")
							sendMsg(fmt.Sprintf(
								"%d %d \nписьмо из группы %d (@%s) от %d (@%s isInUsers: %v): \n%s",
								messID, peer, peer, update.Message.Chat.UserName, from,
								update.Message.From.UserName, isInUsers, msg),
								conf.SupChatID, bot,
							)
							replyToMsg(messID, "Письмо отправлено! Скоро (или нет) придёт ответ", peer, bot)
						} else {
							r := update.Message.ReplyToMessage
							sendMsg(fmt.Sprintf(
								"%d %d \nписьмо из группы (%d @%s) от %d (@%s isInUsers: %v), отвечающее на: \n%s\n(id:%d fr:%d @%s) с текстом:\n%s",
								messID, peer, peer, update.Message.Chat.UserName,
								from, update.Message.From.UserName,
								isInUsers, r.Text, r.MessageID, r.From.ID, r.From.UserName,
								txt), conf.SupChatID, bot,
							)
							replyToMsg(messID, "Письмо отправлено! Скоро (или нет) придёт ответ", peer, bot)
						}
					}
				} else if strings.HasPrefix(strings.ToLower(txt), "атака") {
					aargs := strings.Fields(strings.ToLower(txt))
					if len(aargs) < 2 {
						sendMsg("Атака: аргументов должно быть больше одного", peer, bot)
						return
					}
					args := aargs[1:]
					al := len(args)
					switch args[0] {
					case "статус":
						var ID int64
						if al == 1 {
							if !isInUsers {
								replyToMsg(messID, "Но у вас вомбата нет...", peer, bot)
								return
							}
							ID = int64(from)
						} else if al > 2 {
							replyToMsg(messID, "Атака статус: слишком много аргументов", peer, bot)
							return
						} else {
							strID := args[1]
							if wid, err := strconv.ParseInt(strID, 10, 64); err == nil {
								rCount, err = users.CountDocuments(ctx, bson.M{"_id": wid})
								if err != nil {
									replyToMsg(messID, errStart+"attack: to: count_to", peer, bot)
									rlog.Println("Error: ", err)
									return
								}
								if rCount == 0 {
									replyToMsg(messID, fmt.Sprintf("Ошибка: пользователя с ID %d не существует", wid), peer, bot)
									return
								}
								ID = wid
							} else if wid, ok := womb.Subs[strID]; ok {
								ID = wid
							} else if !ok {
								if len([]rune(strID)) > 64 {
									replyToMsg(messID, "Слишком длинный алиас...", peer, bot)
									return
								}
								replyToMsg(messID, fmt.Sprintf("Ошибка: подписчика с алиасом `%s` не найдено", strID), peer, bot)
								return
							} else {
								replyToMsg(messID, errStart+"attack: to: else", peer, bot)
								rlog.Println("Error: ", "attack: to: else")
								return
							}
						}
						var at Attack
						if is, isFrom := isInAttacks(ID, attacks); isFrom {
							a, err := getAttackByWomb(ID, true, attacks)
							if err != nil {
								replyToMsg(messID, errStart+"attack: status: to_at", peer, bot)
								rlog.Println("Error: ", err)
								return
							}
							at = a
						} else if is {
							a, err := getAttackByWomb(from, false, attacks)
							if err != nil {
								replyToMsg(messID, errStart+"attack: status: from_at", peer, bot)
								rlog.Println("Error: ", err)
								return
							}
							at = a
						} else {
							replyToMsg(messID, "Атак нет", peer, bot)
							return
						}
						var fromWomb, toWomb User
						err = users.FindOne(ctx, bson.M{"_id": at.From}).Decode(&fromWomb)
						if err != nil {
							replyToMsg(messID, errStart+"attack: status: find_fromWomb", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						err = users.FindOne(ctx, bson.M{"_id": at.To}).Decode(&toWomb)
						if err != nil {
							replyToMsg(messID, errStart+"attack: status: finf_towomb", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						replyToMsg(messID, fmt.Sprintf(
							"От: %s (%d)\nКому: %s (%d)\n",
							fromWomb.Name, fromWomb.ID,
							toWomb.Name, toWomb.ID,
						), peer, bot)
					case "атака":
						replyToMsg(messID, strings.Repeat("атака ", 42), peer, bot)
					default:
						replyToMsg(messID, "В группах работает только `статус` и `атака`...", peer, bot)
					}
				} else if isPrefixInList(txt, []string{"рейтинг", "топ"}) {
					args := strings.Fields(strings.ToLower(txt))
					if args[0] != "рейтинг" && args[0] != "топ" {
						return
					}
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
							replyToMsg(messID, fmt.Sprintf("не понимаю, что значит %s", args[1]), peer, bot)
							return
						}
						if len(args) == 3 {
							if isInList(args[2], []string{"+", "плюс", "++", "увеличение"}) {
								queue = 1
							} else if isInList(args[2], []string{"-", "минус", "--", "уменьшение"}) {
								queue = -1
							} else {
								replyToMsg(messID, fmt.Sprintf("не понимаю, что значит %s", args[2]), peer, bot)
								return
							}
						}
					} else if len(args) != 1 {
						replyToMsg(messID, "слишком много аргументов", peer, bot)
						return
					}
					opts := options.Find()
					opts.SetSort(bson.M{name: queue})
					opts.SetLimit(10)
					cur, err := users.Find(ctx, bson.M{}, opts)
					if err != nil {
						replyToMsg(messID, errStart+"rating: find", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
					var rating []User
					for cur.Next(ctx) {
						var w User
						cur.Decode(&w)
						rating = append(rating, w)
					}
					var msg string = fmt.Sprintf("Топ-10 вомбатов по ")
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
						replyToMsg(messID, errStart+"rating: queue else", peer, bot)
						rlog.Println("Error: rating: queue else")
						return
					}
					msg += "\n"
					for num, w := range rating {
						switch name {
						case "money":
							msg += fmt.Sprintf("%d | %s (ID: %d) | %d шишей при себе\n", num+1, w.Name, w.ID, w.Money)
						case "xp":
							msg += fmt.Sprintf("%d | %s (ID: %d) | %d XP\n", num+1, w.Name, w.ID, w.XP)
						case "health":
							msg += fmt.Sprintf("%d | %s (ID: %d) | %d здоровья\n", num+1, w.Name, w.ID, w.Health)
						case "force":
							msg += fmt.Sprintf("%d | %s (ID: %d) | %d мощи\n", num+1, w.Name, w.ID, w.Force)
						}
					}
					msg = strings.TrimSuffix(msg, "\n")
					replyToMsg(messID, msg, peer, bot)
				}
			}(update, titles, bot, users, titlesC, attacks, imgsC, bank)
			continue
		}
		// MESSAGE_DIRECT_CHAT
		go func(update tg.Update, titles []Title, bot *tg.BotAPI, users, titlesC,
			attacks, imgsC, bank *mongo.Collection) {

			peer, from := update.Message.Chat.ID, update.Message.From.ID
			txt, messID := strings.TrimSpace(update.Message.Text), update.Message.MessageID

			const errStart string = "Ошибка... Ответьте командой /admin на это сообщение\n"

			womb := User{}

			wFil := bson.M{"_id": from}

			rCount, err := users.CountDocuments(ctx, wFil)
			checkerr(err)
			if err != nil {
				replyToMsg(messID, errStart+"isInUsers: count_womb", peer, bot)
				rlog.Println("Error: ", err)
				return
			}
			isInUsers := rCount != 0
			if isInUsers {
				err = users.FindOne(ctx, wFil).Decode(&womb)
				if err != nil {
					replyToMsg(messID, errStart+"womb: find_womb", peer, bot)
					rlog.Println("Error: ", err)
					return
				}
			}

			rlog.Printf("MESSAGE p:%d f:%d un:%s, wn:%s, t:%s\n", peer, from, update.Message.From.UserName, womb.Name, txt)

			if isInList(txt, []string{"старт", "начать", "/старт", "/start", "/start@wombatobot", "start", "привет"}) {
				if isInUsers {
					sendMsg(fmt.Sprintf("Здравствуйте, %s!", womb.Name), peer, bot)
				} else {
					sendMsg("Привет! \n — Завести вомбата: `взять вомбата`\n — Помощь: https://telegra.ph/Pomoshch-10-28 (/help)\n — Канал бота, где есть нужная инфа: @wombatobot_channel\n Приятной игры!",
						peer, bot,
					)
				}
			} else if isInList(txt, []string{"взять вомбата", "купить вомбата у арабов", "хочу вомбата"}) {
				if isInUsers {
					sendMsg("У тебя как бы уже есть вомбат лолкек. Если хочешь от него избавиться, то напиши `приготовить шашлык`",
						peer, bot,
					)
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
						Sleep:  false,
					}
					_, err = users.InsertOne(ctx, &newWomb)
					if err != nil {
						replyToMsg(messID, errStart+"new_womb: insert", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
					newimg, err := getImgs(imgsC, "new")
					if err != nil {
						replyToMsg(messID, errStart+"new_womb: get_imgs", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
					sendPhoto(randImg(newimg), fmt.Sprintf(
						"Поздравляю, у тебя появился вомбат! Ему выдалось имя `%s`. Ты можешь поменять имя командой `Поменять имя [имя]` за 3 монеты",
						newWomb.Name),
						peer, bot,
					)
				}
			} else if strings.HasPrefix(strings.ToLower(txt), "devtools") {
				if hasTitle(0, womb.Titles) {
					cmd := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(txt), "devtools"))
					if strings.HasPrefix(cmd, "set money") {
						strNewMoney := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(cmd), "set money"))
						if i, err := strconv.ParseUint(strNewMoney, 10, 64); err == nil {
							womb.Money = i
							err = docUpd(womb, wFil, users)
							if err != nil {
								replyToMsg(messID, errStart+"devtools: set_money: upd", peer, bot)
								rlog.Println("Error: ", err)
								return
							}
							sendMsg(fmt.Sprintf("Операция проведена успешно! Шишей при себе: %d", womb.Money), peer, bot)
						} else {
							sendMsg("Ошибка: неправильный синтаксис. Синтаксис команды: `devtools set money {кол-во шишей}`", peer, bot)
						}
					} else if strings.HasPrefix(cmd, "reset") {
						arg := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(cmd), "reset"))
						switch arg {
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
							sendMsg("Ошибка: неправильный синтаксис. Синтаксис команды: `devtools reset [force/health/xp/all]`",
								peer, bot,
							)
							return
						}
						err := docUpd(womb, wFil, users)
						if err != nil {
							replyToMsg(messID, errStart+"devtools: reset: update", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						sendMsg("Операция произведена успешно!", peer, bot)
					} else if cmd == "help" {
						sendMsg("https://telegra.ph/Vombot-devtools-help-10-28", peer, bot)
					}
				} else if strings.TrimSpace(txt) == "devtools on" {
					womb.Titles = append(womb.Titles, 0)
					err := docUpd(womb, wFil, users)
					if err != nil {
						replyToMsg(messID, errStart+"devtools: on", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
					sendMsg("Выдан титул \"Вомботестер\" (ID: 0)", peer, bot)
				}
			} else if isInList(txt, []string{"приготовить шашлык", "продать вомбата арабам", "слить вомбата в унитаз", "убить"}) {
				if isInUsers {
					if !(hasTitle(1, womb.Titles)) {
						_, err = users.DeleteOne(ctx, wFil)
						if err != nil {
							replyToMsg(messID, errStart+"kill: delete", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						kill, err := getImgs(imgsC, "kill")
						if err != nil {
							replyToMsg(messID, errStart+"kill: get_imgs", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						sendPhoto(randImg(kill), "Вы уничтожили вомбата в количестве 1 штука. Вы - нехорошее существо", peer, bot)
					} else {
						sendMsg("Ошибка: вы лишены права уничтожать вомбата; ответьте на это сообщение командой /admin для объяснений",
							peer, bot)
					}
				} else {
					sendMsg("Но у вас нет вомбата...", peer, bot)
				}
			} else if strings.HasPrefix(strings.ToLower(txt), "поменять имя") {
				if isInUsers {
					if hasTitle(1, womb.Titles) {
						sendMsg("Тебе нельзя, ты спамер (оспорить: /admin)", peer, bot)
						return
					}
					name := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(txt), "поменять имя"))
					if womb.Money >= 3 {
						if isInList(name, []string{"admin", "вoмбoт", "вoмбoт", "вомбoт", "вомбот", "бот", "bot", "бoт", "bоt",
							"авто", "auto"}) {
							sendMsg("Такие никнеймы заводить нельзя", peer, bot)
						} else if name != "" {
							if len([]rune(name)) > 64 {
								replyToMsg(messID, "Слишком длинный никнейм!", peer, bot)
								return
							}
							womb.Money -= 3
							split := strings.Fields(txt)
							caseName := strings.Join(split[2:], " ")
							womb.Name = caseName
							err := docUpd(womb, wFil, users)
							if err != nil {
								replyToMsg(messID, errStart+"rename: update", peer, bot)
								rlog.Println("Error: ", err)
								return
							}
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
						if uint64(womb.Health+1) < uint64(math.Pow(2, 32)) {
							womb.Money -= 5
							womb.Health++
							err := docUpd(womb, wFil, users)
							if err != nil {
								replyToMsg(messID, errStart+"buy_health: update", peer, bot)
								rlog.Println("Error: ", err)
								return
							}
							sendMsg(fmt.Sprintf("Поздравляю! Теперь у вас %d здоровья и %d шишей при себе", womb.Health, womb.Money), peer, bot)
						} else {
							sendMsg(
								"Ошибка: вы достигли максимального количества здоровья (2 в 32 степени). Если это вас возмущает, ответьте командой /admin",
								peer, bot,
							)
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
						if uint64(womb.Force+1) < uint64(math.Pow(2, 32)) {
							womb.Money -= 3
							womb.Force++
							err := docUpd(womb, wFil, users)
							if err != nil {
								replyToMsg(messID, errStart+"buy_force: update", peer, bot)
								rlog.Println("Error: ", err)
								return
							}
							sendMsg(fmt.Sprintf("Поздравляю! Теперь у вас %d мощи и %d шишей при себе", womb.Force, womb.Money), peer, bot)
						} else {
							sendMsg(
								"Ошибка: вы достигли максимального количества мощи (2 в 32 степени). Если это вас возмущает, ответьте командой /admin",
								peer, bot,
							)
						}
					} else {
						sendMsg("Надо накопить побольше шишей! 1 мощь = 3 шиша", peer, bot)
					}
				} else {
					sendMsg("У тя ваще вобата нет...", peer, bot)
				}
			} else if isInList(txt, []string{"поиск денег"}) {
				if isInUsers {
					if womb.Money < 5 {
						womb.Money = 5
						err := docUpd(womb, wFil, users)
						if err != nil {
							replyToMsg(messID, errStart+"find_money: free", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						replyToMsg(messID, "Так как у вас было меньше 5 шишей при себе, у вас их теперь 5!", peer, bot)
						return
					}
					if womb.Money >= 1 {
						womb.Money--
						rand.Seed(time.Now().UnixNano())
						if ch := rand.Int(); ch%2 == 0 || hasTitle(2, womb.Titles) && (ch%2 == 0 || ch%3 == 0) {
							rand.Seed(time.Now().UnixNano())
							win := rand.Intn(9) + 1
							womb.Money += uint64(win)
							if addXP := rand.Intn(512 - 1); addXP < 5 {
								womb.XP += uint32(addXP)
								sendMsg(fmt.Sprintf(
									"Поздравляем! Вы нашли на дороге %d шишей, а ещё вам дали %d XP! Теперь у вас %d шишей при себе и %d XP",
									win, addXP, womb.Money, womb.XP),
									peer, bot,
								)
							} else {
								sendMsg(fmt.Sprintf("Поздравляем! Вы нашли на дороге %d шишей! Теперь их у вас при себе %d", win, womb.Money),
									peer, bot,
								)
							}
						} else {
							sendMsg("Вы заплатили один шиш охранникам денежной дорожки, но увы, вы так ничего и не нашли", peer, bot)
						}
						err := docUpd(womb, wFil, users)
						if err != nil {
							replyToMsg(messID, errStart+"find_money: update", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
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
					ID := uint16(i)
					rCount, err := titlesC.CountDocuments(ctx, bson.M{"_id": ID})
					if err != nil {
						replyToMsg(messID, errStart+"about_title: count_titles", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
					if rCount != 0 {
						elem := Title{}
						err = titlesC.FindOne(ctx, bson.M{"_id": ID}).Decode(&elem)
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
					sendMsg("Ошибка: пропущены аргументы `ID` и `алиас`. Синтаксис команды: `подписаться [ID] [алиас (без пробелов)]`",
						peer, bot,
					)
				} else if len(args) == 1 {
					sendMsg("Ошибка: пропущен аргумент `алиас`. Синтаксис команды: `подписаться [ID] [алиас (без пробелов)]`", peer, bot)
				} else if len(args) == 2 {
					if ID, err := strconv.ParseInt(args[0], 10, 64); err == nil {
						if _, err := strconv.ParseInt(args[1], 10, 64); err == nil {
							sendMsg("Ошибка: алиас не должен быть числом", peer, bot)
						} else if len([]rune(args[1])) > 64 {
							sendMsg("Слишком длинный алиас!", peer, bot)
						} else {
							if elem, ok := womb.Subs[args[1]]; !ok {
								rCount, err = users.CountDocuments(ctx, bson.M{"_id": ID})
								if err != nil {
									replyToMsg(messID, errStart+"subscribe: count", peer, bot)
									rlog.Println("Error: ", err)
									return
								}
								subbed, name := isInSubs(ID, womb.Subs)
								if subbed {
									sendMsg(fmt.Sprintf(
										"Ошибка: вы уже подписались на вомбата с ID %d (алиас: %s). Для того, чтобы отписаться, напишите команду \"отписаться %s\"",
										ID, name, name),
										peer, bot,
									)
									return
								}
								if rCount != 0 {
									womb.Subs[args[1]] = ID
									err := docUpd(womb, wFil, users)
									if err != nil {
										replyToMsg(messID, errStart+"subscribe: update", peer, bot)
										rlog.Println("Error: ", err)
										return
									}
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
				if len([]rune(alias)) > 64 {
					replyToMsg(messID, "Слишком длинный алиас...", peer, bot)
					return
				}
				if _, ok := womb.Subs[alias]; ok {
					delete(womb.Subs, alias)
					err := docUpd(womb, wFil, users)
					if err != nil {
						replyToMsg(messID, errStart+"unsub: update", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
					sendMsg(fmt.Sprintf("Вы отписались от пользователя с алиасом %s", alias), peer, bot)
				} else {
					sendMsg(fmt.Sprintf("Ошибка: вы не подписаны на пользователя с алиасом `%s`", alias), peer, bot)
				}
			} else if isInList(txt, []string{"подписки", "мои подписки", "список подписок"}) {
				strSubs := "Вот список твоих подписок:"
				if len(womb.Subs) != 0 {
					for alias, id := range womb.Subs {
						rCount, err = users.CountDocuments(ctx, bson.M{"_id": id})
						if err != nil {
							replyToMsg(messID, errStart+"sub_list: count", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						if rCount != 0 {
							tWomb := User{}
							err = users.FindOne(ctx, bson.M{"_id": id}).Decode(&tWomb)
							if err != nil {
								replyToMsg(messID, errStart+"sub_list: find", peer, bot)
								rlog.Println("Error: ", err)
								return
							}
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
					rCount, err := users.CountDocuments(ctx, bson.M{"_id": ID})
					if err != nil {
						replyToMsg(messID, errStart+"list_subs: count", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
					if rCount != 0 {
						tWomb := User{}
						err = users.FindOne(ctx, bson.M{"_id": ID}).Decode(&tWomb)
						if err != nil {
							replyToMsg(messID, errStart+"list_subs: find", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						strTitles := ""
						tCount := len(tWomb.Titles)
						if tCount != 0 {
							for _, id := range tWomb.Titles {
								rCount, err = titlesC.CountDocuments(ctx, bson.M{"_id": id})
								if err != nil {
									replyToMsg(messID, errStart+"list_subs: count_titles", peer, bot)
									rlog.Println("Error: ", err)
									return
								}
								if rCount == 0 {
									strTitles += fmt.Sprintf("Ошибка: титула с ID %d нет (ответьте командой /admin) |", id)
									continue
								}
								elem := Title{}
								err = titlesC.FindOne(ctx, bson.M{"_id": id}).Decode(&elem)
								if err != nil {
									replyToMsg(messID, errStart+"list_subs: find_titles", peer, bot)
									rlog.Println("Error: ", err)
									return
								}
								strTitles += fmt.Sprintf("%s (ID: %d) | ", elem.Name, id)
							}
							strTitles = strings.TrimSuffix(strTitles, " | ")
						} else {
							strTitles = "нет"
						}
						sendMsg(fmt.Sprintf(
							"Вомбат  %s (ID: %d; Алиас: %s)\nТитулы: %s\n 🕳 %d XP \n ❤ %d здоровья \n ⚡ %d мощи \n 💰 %d шишей при себе",
							tWomb.Name, ID, alias, strTitles, tWomb.XP, tWomb.Health, tWomb.Force, tWomb.Money), peer, bot,
						)
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
					rCount, err = users.CountDocuments(ctx, bson.M{"_id": ID})
					if err != nil {
						replyToMsg(messID, errStart+"about_womb: id: count", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
					if rCount == 0 {
						sendMsg(fmt.Sprintf("Ошибка: пользователя с ID %d не существует", ID), peer, bot)
						return
					}
					err = users.FindOne(ctx, bson.M{"_id": ID}).Decode(&tWomb)
					if err != nil {
						replyToMsg(messID, errStart+"about_womb: id: find", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
				} else if ID, ok = womb.Subs[strID]; ok {
					err = users.FindOne(ctx, bson.M{"_id": ID}).Decode(&tWomb)
					if err != nil {
						replyToMsg(messID, errStart+"about_womb: alias: find", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
				} else if !ok {
					if len([]rune(strID)) > 64 {
						replyToMsg(messID, "Слишком длинный алиас...", peer, bot)
						return
					}
					replyToMsg(messID, fmt.Sprintf("Ошибка: подписчика с алиасом `%s` не найдено", strID), peer, bot)
					return
				} else {
					replyToMsg(messID, errStart+"about womb: else", peer, bot)
					rlog.Println("Error: about womb: else")
					return
				}
				strTitles := ""
				tCount := len(tWomb.Titles)
				if tCount != 0 {
					for _, id := range tWomb.Titles {
						elem := Title{}
						rCount, err = titlesC.CountDocuments(ctx, bson.M{"_id": id})
						if err != nil {
							replyToMsg(messID, errStart+"list_subs: count_titles", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						if rCount == 0 {
							strTitles += fmt.Sprintf("Ошибка: титула с ID %d нет (ответьте командой /admin) |", id)
							continue
						}
						err = titlesC.FindOne(ctx, bson.M{"_id": id}).Decode(&elem)
						if err != nil {
							replyToMsg(messID, errStart+"about_womb: find_title", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						strTitles += fmt.Sprintf("%s (ID: %d) | ", elem.Name, id)
					}
					strTitles = strings.TrimSuffix(strTitles, " | ")
				} else {
					strTitles = "нет"
				}
				var sl string
				if womb.Sleep {
					sl = "Спит"
				} else {
					sl = "Не спит"
				}
				abimg, err := getImgs(imgsC, "about")
				if err != nil {
					replyToMsg(messID, errStart+"about_womb: get_imgs", peer, bot)
					rlog.Println("Error: ", err)
					return
				}
				sendPhotoMD(randImg(abimg), fmt.Sprintf("Вомбат %s (ID: %d) {%s}\nТитулы: %s\n 🕳 %d XP \n ❤ %d здоровья \n ⚡ %d мощи \n 💰 %d шишей при себе",
					tWomb.Name, ID, sl, strTitles, tWomb.XP, tWomb.Health, tWomb.Force, tWomb.Money), peer, bot,
				)
			} else if strings.HasPrefix(strings.ToLower(txt), "о вомботе") {
				sendMsgMD("https://telegra.ph/O-vombote-10-29\n**если вы хотели узнать характеристики вомбата, используйте команду `о вомбате`**",
					peer, bot,
				)
			} else if strings.HasPrefix(strings.ToLower(txt), "перевести шиши") {
				args := strings.Fields(strings.TrimSpace(strings.TrimPrefix(strings.ToLower(txt), "перевести шиши")))
				if len(args) < 2 {
					sendMsg("Ошибка: вы пропустили аргумент(ы). Синтаксис команды: `перевести шиши [кол-во] [ID/алиас получателя]`",
						peer, bot,
					)
				} else if len(args) > 2 {
					sendMsg("Ошибка: слишком много аргументов. Синтаксис команды: `перевести шиши [кол-во] [ID/алиас получателя]`",
						peer, bot,
					)
				} else {
					if amount, err := strconv.ParseUint(args[0], 10, 64); err == nil {
						var ID int64
						if ID, err = strconv.ParseInt(args[1], 10, 64); err != nil {
							var ok bool
							if ID, ok = womb.Subs[args[1]]; !ok {
								if len([]rune(args[1])) > 64 {
									replyToMsg(messID, "Слишком длинный алиас...", peer, bot)
									return
								}
								sendMsg(fmt.Sprintf("Ошибка: алиаса %s не обнаружено", args[1]), peer, bot)
								return
							}
						}
						if womb.Money >= amount {
							if amount != 0 {
								if ID == peer {
									sendMsg("Ты читер блин нафиг!!!!!! нидам тебе самому себе перевести", peer, bot)
									return
								}
								rCount, err = users.CountDocuments(ctx, bson.M{"_id": ID})
								if err != nil {
									replyToMsg(messID, errStart+"send_shishs: count_to", peer, bot)
									rlog.Println("Error: ", err)
									return
								}
								if rCount != 0 {
									tWomb := User{}
									err = users.FindOne(ctx, bson.M{"_id": ID}).Decode(&tWomb)
									if err != nil {
										replyToMsg(messID, errStart+"send_shishs: find_to", peer, bot)
										rlog.Println("Error: ", err)
										return
									}
									womb.Money -= amount
									tWomb.Money += amount
									err := docUpd(tWomb, bson.M{"_id": ID}, users)
									if err != nil {
										replyToMsg(messID, errStart+"send_shishs: update: from", peer, bot)
										rlog.Println("Error: ", err)
										return
									}
									err = docUpd(womb, wFil, users)
									if err != nil {
										replyToMsg(messID, errStart+"send_shishs: update: to", peer, bot)
										rlog.Println("Error: ", err)
										return
									}
									sendMsg(fmt.Sprintf("Вы успешно перевели %d шишей на счёт %s. Теперь у вас %d шишей при себе",
										amount, tWomb.Name, womb.Money), peer, bot,
									)
									sendMsg(fmt.Sprintf("Пользователь %s (ID: %d) перевёл вам %d шишей. Теперь у вас %d шишей при себе",
										womb.Name, peer, amount, tWomb.Money), ID, bot,
									)
								} else {
									sendMsg(fmt.Sprintf("Ошибка: пользователя с ID %d не найдено", ID), peer, bot)
								}
							} else {
								sendMsg("Ошибка: количество переводимых шишей должно быть больше нуля", peer, bot)
							}
						} else {
							sendMsg(fmt.Sprintf("Ошибка: размер перевода (%d) должен быть меньше кол-ва ваших шишей при себе (%d)",
								amount, womb.Money), peer, bot,
							)
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
				users = db.Collection("users")
				attacks = db.Collection("attacks")
				titlesC := db.Collection("titles")
				cur, err := titlesC.Find(ctx, bson.M{})
				defer cur.Close(ctx)
				if err != nil {
					replyToMsg(messID, errStart+"update_data: titles", peer, bot)
					rlog.Println("Error: ", err)
					return
				}
				titles = []Title{}
				for cur.Next(ctx) {
					var nextOne Title
					err := cur.Decode(&nextOne)
					if err != nil {
						replyToMsg(messID, errStart+"update_data: titles_decode", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
					titles = append(titles, nextOne)
				}
				cur.Close(ctx)
				imgsC = db.Collection("imgs")
				sendMsg("Успешно обновлено!", peer, bot)
				rlog.Printf("DATA_UPDATE %d\n", peer)
				fmt.Printf("Data update by %d\n", peer)
			} else if isInList(txt, []string{"купить квес", "купить квесс", "купить qwess", "попить квес", "попить квесс", "попить qwess"}) {
				if isInUsers {
					if womb.Money >= 256 {
						qwess, err := getImgs(imgsC, "qwess")
						if err != nil {
							replyToMsg(messID, errStart+"nyamka: get_qwess_imgs", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						if !(hasTitle(2, womb.Titles)) {
							womb.Titles = append(womb.Titles, 2)
							womb.Money -= 256
							err = docUpd(womb, wFil, users)
							if err != nil {
								replyToMsg(messID, errStart+"nyamka: update_first_time", peer, bot)
								rlog.Println("Error: ", err)
								return
							}
							sendPhoto(
								randImg(qwess),
								"Вы купили чудесного вкуса квес у кролика-Лепса в ларьке за 256 шишей. Глотнув этот напиток, вы поняли, что получили новый титул с ID 2",
								peer, bot,
							)
						} else {
							womb.Money -= 256
							err = docUpd(womb, wFil, users)
							if err != nil {
								replyToMsg(messID, errStart+"nyamka: update", peer, bot)
								rlog.Println("Error: ", err)
								return
							}
							if err != nil {
								replyToMsg(messID, errStart+"nyamka: get_leps_imgs", peer, bot)
								rlog.Println("Error: ", err)
								return
							}
							sendPhoto(
								randImg(qwess),
								"Вы вновь купили вкусного квеса у того же кролика-Лепса в том же ларьке за 256 шишей. \"Он так освежает, я чувствую себя человеком\" — думаете вы. Ах, как вкусён квес!",
								peer, bot,
							)
						}
					} else {
						leps, err := getImgs(imgsC, "leps")
						if err != nil {
							replyToMsg(messID, errStart+"nyamka: get_leps_imgs", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						sendPhoto(
							randImg(leps),
							"Вы подошли к ближайшему ларьку, но, увы, кролик-Лепс на кассе сказал, что надо 256 шишей, а у вас, к сожалению, меньше",
							peer, bot,
						)
					}
				} else {
					sendMsg("К сожалению, вам нужны шиши, чтобы купить квес, а шиши есть только у вомбатов...", peer, bot)
				}
			} else if isPrefixInList(txt, []string{"/admin", "/админ", "/admin@wombatobot", "одмен!", "/баг", "/bug", "/bug@wombatobot", "/support", "/support@wombatobot"}) {
				oArgs := strings.Fields(strings.ToLower(txt))
				if len(oArgs) < 2 {
					if update.Message.ReplyToMessage == nil {
						replyToMsg(messID, "Ты чаво... где письмо??", peer, bot)
						return
					}
					r := update.Message.ReplyToMessage
					sendMsg(fmt.Sprintf(
						"%d %d \nписьмо от %d (@%s isInUsers: %v), отвечающее на: \n%s\n(id:%d fr:%d @%s)",
						messID, peer, from, update.Message.From.UserName,
						isInUsers, r.Text, r.MessageID, r.From.ID, r.From.UserName),
						conf.SupChatID, bot,
					)
					replyToMsg(messID, "Письмо отправлено! Скоро (или нет) придёт ответ", peer, bot)
				} else {
					if update.Message.ReplyToMessage == nil {
						msg := strings.Join(oArgs[1:], " ")
						sendMsg(fmt.Sprintf(
							"%d %d \nписьмо %d (@%s) от %d (@%s isInUsers: %v): \n%s",
							messID, peer, peer, update.Message.Chat.UserName, from,
							update.Message.From.UserName, isInUsers, msg),
							conf.SupChatID, bot,
						)
						replyToMsg(messID, "Письмо отправлено! Скоро (или нет) придёт ответ", peer, bot)
					} else {
						r := update.Message.ReplyToMessage
						sendMsg(fmt.Sprintf(
							"%d %d \nписьмо от %d (@%s isInUsers: %v), отвечающее на: \n%s\n(id:%d fr:%d @%s) с текстом:\n%s",
							messID, peer, from, update.Message.From.UserName,
							isInUsers, r.Text, r.MessageID, r.From.ID, r.From.UserName,
							txt), conf.SupChatID, bot,
						)
						replyToMsg(messID, "Письмо отправлено! Скоро (или нет) придёт ответ", peer, bot)
					}
				}
			} else if strings.HasPrefix(strings.ToLower(txt), "атака") {
				aargs := strings.Fields(strings.ToLower(txt))
				if len(aargs) < 2 {
					sendMsg("Атака: аргументов должно быть больше одного", peer, bot)
					return
				}
				args := aargs[1:]
				al := len(args)
				switch args[0] {
				case "атака":
					sendMsg(strings.Repeat("атака ", 42), peer, bot)
				case "на":
					if al < 2 {
						sendMsg("Атака на: на кого?", peer, bot)
						return
					} else if al != 2 {
						sendMsg("Атака на: слишком много аргументов", peer, bot)
						return
					} else if !isInUsers {
						sendMsg("Вы не можете атаковать в виду остутствия вомбата", peer, bot)
						return
					} else if womb.Sleep {
						sendMsg("Но вы же спите...", peer, bot)
						return
					}
					strID := args[1]
					var (
						ID    int64
						tWomb User
						ok    bool
					)
					if is, isFrom := isInAttacks(from, attacks); isFrom {
						at, err := getAttackByWomb(from, true, attacks)
						if err != nil && err != errNoAttack {
							replyToMsg(messID, errStart+"attack: to: from_from: get_attack_by_womb", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						var aWomb User
						err = users.FindOne(ctx, bson.M{"_id": at.To}).Decode(&aWomb)
						if err != nil {
							replyToMsg(messID, errStart+"attack: to: find_attack_from", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						sendMsgMD(fmt.Sprintf(
							"Вы уже атакуете вомбата %s (ID: %d). Чтобы отозвать атаку, напишите `атака отмена`",
							aWomb.Name, aWomb.ID),
							peer, bot)
						return
					} else if is {
						at, err := getAttackByWomb(from, false, attacks)
						if err != nil && err != errNoAttack {
							replyToMsg(messID, errStart+"attack: to: from_to: get_attack_by_womb", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						var aWomb User
						err = users.FindOne(ctx, bson.M{"_id": at.From}).Decode(&aWomb)
						if err != nil {
							replyToMsg(messID, errStart+"attack: to: find_attack_to", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						sendMsgMD(fmt.Sprintf(
							"Вас уже атакует вомбат %s (ID: %d). Чтобы отклонить атаку, напишите `атака отмена`",
							aWomb.Name, aWomb.ID),
							peer, bot)
						return
					}
					if ID, err = strconv.ParseInt(strID, 10, 64); err == nil {
						rCount, err = users.CountDocuments(ctx, bson.M{"_id": ID})
						if err != nil {
							replyToMsg(messID, errStart+"attack: to: count_to", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						if rCount == 0 {
							sendMsg(fmt.Sprintf("Ошибка: пользователя с ID %d не существует", ID), peer, bot)
							return
						}
						err = users.FindOne(ctx, bson.M{"_id": ID}).Decode(&tWomb)
						if err != nil {
							replyToMsg(messID, errStart+"attack: to: find_to", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
					} else if ID, ok = womb.Subs[strID]; ok {
						err = users.FindOne(ctx, bson.M{"_id": womb.Subs[strID]}).Decode(&tWomb)
						if err != nil {
							replyToMsg(messID, errStart+"attack: to ", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
					} else if !ok {
						if len([]rune(strID)) > 64 {
							replyToMsg(messID, "Слишком длинный алиас...", peer, bot)
							return
						}
						replyToMsg(messID, fmt.Sprintf("Ошибка: подписчика с алиасом `%s` не найдено", strID), peer, bot)
						return
					} else {
						replyToMsg(messID, errStart+"attack: to: else", peer, bot)
						rlog.Println("Error: ", "attack: to: else")
						return
					}
					if ID == int64(from) {
						sendMsg("„Основная борьба в нашей жизни — борьба с самим собой“ (c) какой-то философ", peer, bot)
						return
					}
					err = users.FindOne(ctx, bson.M{"_id": ID}).Decode(&tWomb)
					if err != nil {
						replyToMsg(messID, errStart+"attack: to: is_to_sleep", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
					if tWomb.Sleep {
						sendMsg(fmt.Sprintf(
							"Вомбат %s спит. Его атаковать не получится",
							tWomb.Name), peer, bot)
						return
					} else if is, isFrom := isInAttacks(ID, attacks); isFrom {
						at, err := getAttackByWomb(ID, true, attacks)
						if err != nil && err != errNoAttack {
							replyToMsg(messID, errStart+"attack: to: to_from: get_attack_by_womb", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						var aWomb User
						err = users.FindOne(ctx, bson.M{"_id": at.To}).Decode(&aWomb)
						if err != nil {
							replyToMsg(messID, errStart+"attack: to: find_to_from", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						sendMsgMD(fmt.Sprintf(
							"%s уже атакует вомбата %s (ID: %d). Попросите %s решить данную проблему",
							strID, aWomb.Name, aWomb.ID, strID),
							peer, bot)
						return
					} else if is {
						at, err := getAttackByWomb(from, false, attacks)
						if err != nil && err != errNoAttack {
							replyToMsg(messID, errStart+"attack: to: to_to: get_attack_by_womb", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						var aWomb User
						err = users.FindOne(ctx, bson.M{"_id": at.From}).Decode(&aWomb)
						if err != nil {
							replyToMsg(messID, errStart+"attack: to: find_to_to", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						sendMsgMD(fmt.Sprintf(
							"Вомбат %s уже атакуется %s (ID: %d). Попросите %s решить данную проблему",
							strID, aWomb.Name, aWomb.ID, strID),
							peer, bot)
						return
					}
					var newAt = Attack{
						ID:   strconv.Itoa(int(from)) + "_" + strconv.Itoa(int(ID)),
						From: int64(from),
						To:   ID,
					}
					_, err = attacks.InsertOne(ctx, newAt)
					if err != nil {
						replyToMsg(messID, errStart+"attack: to: insert", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
					sendMsg(fmt.Sprintf(
						"Вы отправили вомбата атаковать %s. Ждём ответа!\nОтменить можно командой `атака отмена`",
						tWomb.Name), peer, bot)
					sendMsg(fmt.Sprintf(
						"Ужас! Вас атакует %s. Предпримите какие-нибудь меры: отмените атаку (`атака отмена`) или примите (`атака принять`)",
						womb.Name), tWomb.ID, bot)
				case "статус":
					var ID int64
					if al == 1 {
						if !isInUsers {
							sendMsg("Но у вас вомбата нет...", peer, bot)
							return
						}
						ID = int64(from)
					} else if al > 2 {
						sendMsg("Атака статус: слишком много аргументов", peer, bot)
						return
					} else {
						strID := args[1]
						if wid, err := strconv.ParseInt(strID, 10, 64); err == nil {
							rCount, err = users.CountDocuments(ctx, bson.M{"_id": wid})
							if err != nil {
								replyToMsg(messID, errStart+"attack: to: count_to", peer, bot)
								rlog.Println("Error: ", err)
								return
							}
							if rCount == 0 {
								sendMsg(fmt.Sprintf("Ошибка: пользователя с ID %d не существует", wid), peer, bot)
								return
							}
							ID = wid
						} else if wid, ok := womb.Subs[strID]; ok {
							ID = wid
						} else if !ok {
							if len([]rune(strID)) > 64 {
								replyToMsg(messID, "Слишком длинный алиас...", peer, bot)
								return
							}
							replyToMsg(messID, fmt.Sprintf("Ошибка: подписчика с алиасом `%s` не найдено", strID), peer, bot)
							return
						} else {
							replyToMsg(messID, errStart+"attack: to: else", peer, bot)
							rlog.Println("Error: ", "attack: to: else")
							return
						}
					}
					var at Attack
					if is, isFrom := isInAttacks(ID, attacks); isFrom {
						a, err := getAttackByWomb(ID, true, attacks)
						if err != nil {
							replyToMsg(messID, errStart+"attack: status: to_at", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						at = a
					} else if is {
						a, err := getAttackByWomb(from, false, attacks)
						if err != nil {
							replyToMsg(messID, errStart+"attack: status: from_at", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						at = a
					} else {
						sendMsg("Атак нет", peer, bot)
						return
					}
					var fromWomb, toWomb User
					err = users.FindOne(ctx, bson.M{"_id": at.From}).Decode(&fromWomb)
					if err != nil {
						replyToMsg(messID, errStart+"attack: status: find_fromWomb", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
					err = users.FindOne(ctx, bson.M{"_id": at.To}).Decode(&toWomb)
					if err != nil {
						replyToMsg(messID, errStart+"attack: status: finf_towomb", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
					sendMsg(fmt.Sprintf(
						"От: %s (%d)\nКому: %s (%d)\n",
						fromWomb.Name, fromWomb.ID,
						toWomb.Name, toWomb.ID,
					), peer, bot)
				case "отмена":
					if al > 1 {
						sendMsg("атака отмена: слишком много аргументов", peer, bot)
					} else if !isInUsers {
						sendMsg("какая атака, у тебя вобмата нет", peer, bot)
					}
					var at Attack
					if is, isFrom := isInAttacks(from, attacks); isFrom {
						a, err := getAttackByWomb(from, true, attacks)
						if err != nil {
							replyToMsg(messID, errStart+"attack: cancel: to_at", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						at = a
					} else if is {
						a, err := getAttackByWomb(from, false, attacks)
						if err != nil {
							replyToMsg(messID, errStart+"attack: cancel: from_at", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						at = a
					} else {
						sendMsg("Атаки с вами не найдено...", peer, bot)
						return
					}
					_, err = attacks.DeleteOne(ctx, bson.M{"_id": at.ID})
					if err != nil {
						replyToMsg(messID, errStart+"attack: cancel: delete", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
					can0, err := getImgs(imgsC, "cancel_0")
					if err != nil {
						replyToMsg(messID, errStart+"attack: cancel: get_imgs_0", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
					can1, err := getImgs(imgsC, "cancel_1")
					if err != nil {
						replyToMsg(messID, errStart+"attack: cancel: get_imgs_1", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
					if at.From == int64(from) {
						sendPhoto(randImg(can0), "Вы отменили атаку", peer, bot)
						sendPhoto(randImg(can1),
							fmt.Sprintf("Вомбат %s решил вернуть вомбата домой. Вы свободны от атак",
								womb.Name), at.To, bot)
					} else {
						sendPhoto(randImg(can0), "Вы отклонили атаку", peer, bot)
						sendPhoto(randImg(can1), fmt.Sprintf(
							"Вомбат %s вежливо отказал вам в войне. Вам пришлось забрать вомбата обратно. Вы свободны от атак",
							womb.Name), at.From, bot)
					}
				case "принять":
					if al > 2 {
						sendMsg("Атака статус: слишком много аргументов", peer, bot)
						return
					} else if !isInUsers {
						sendMsg("Но у вас вомбата нет...", peer, bot)
						return
					}
					var at Attack
					if is, isFrom := isInAttacks(from, attacks); isFrom {
						sendMsg("Ну ты чо... атаку принимает тот, кого атакуют...", peer, bot)
					} else if is {
						a, err := getAttackByWomb(from, false, attacks)
						if err != nil {
							replyToMsg(messID, errStart+"attack: cancel: from_at", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						at = a
					} else {
						sendMsg("Вам нечего принимать...", peer, bot)
						return
					}
					rCount, err = users.CountDocuments(ctx, bson.M{"_id": at.From})
					if err != nil {
						replyToMsg(messID, errStart+"attack: accept: count_from", peer, bot)
						rlog.Println("Error: ", err)
						return
					} else if rCount < 1 {
						sendMsg("Ну ты чаво... Соперника не существует! Как вообще мы такое допустили?! (ответь на это командой /admin",
							peer, bot)
						return
					}
					var tWomb User
					err = users.FindOne(ctx, bson.M{"_id": at.From}).Decode(&tWomb)
					if err != nil {
						replyToMsg(messID, errStart+"attack: accept: find_from", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
					atimgs, err := getImgs(imgsC, "attacks")
					if err != nil {
						replyToMsg(messID, errStart+"attack: accept: imgs", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
					im := randImg(atimgs)
					ph1 := sendPhoto(im, "", peer, bot)
					ph2 := sendPhoto(im, "", tWomb.ID, bot)
					war1 := replyToMsg(ph1, "Да начнётся вомбой!", peer, bot)
					war2 := replyToMsg(ph2, fmt.Sprintf(
						"АААА ВАЙНААААА!!!\n Вомбат %s всё же принял ваше предложение",
						womb.Name), tWomb.ID, bot,
					)
					time.Sleep(5 * time.Second)
					h1, h2 := int(womb.Health), int(tWomb.Health)
					for _, round := range []int{1, 2, 3} {
						f1 := uint32(2 + rand.Intn(int(womb.Force-1)))
						f2 := uint32(2 + rand.Intn(int(tWomb.Force-1)))
						editMsg(war1, fmt.Sprintf(
							"РАУНД %d\n\nВаш вомбат:\n - здоровье: %d\n -Ваш удар: %d\n\n%s:\n - здоровье: %d",
							round, h1, f1, tWomb.Name, h2), peer, bot)
						editMsg(war2, fmt.Sprintf(
							"РАУНД %d\n\nВаш вомбат:\n - здоровье: %d\n - Ваш удар: %d\n\n%s:\n - здоровье: %d",
							round, h2, f2, womb.Name, h1), tWomb.ID, bot)
						time.Sleep(3 * time.Second)
						h1 -= int(f2)
						h2 -= int(f1)
						editMsg(war1, fmt.Sprintf(
							"РАУНД %d\n\nВаш вомбат:\n - здоровье: %d\n - Ваш удар: %d\n\n%s:\n - здоровье: %d\n - 💔 удар: %d",
							round, h1, f1, tWomb.Name, h2, f2), peer, bot)
						editMsg(war2, fmt.Sprintf(
							"РАУНД %d\n\nВаш вомбат:\n - здоровье: %d\n - Ваш удар: %d\n\n%s:\n - здоровье: %d\n - 💔 удар: %d",
							round, h2, f2, womb.Name, h1, f1), tWomb.ID, bot)
						time.Sleep(5 * time.Second)
						if int(h2)-int(f1) <= 5 && int(h1)-int(f2) <= 5 {
							editMsg(war1,
								"Вы оба сдохли!!!)\nВаши характеристики не поменялись, но зато да.",
								peer, bot)
							editMsg(war2,
								"Вы оба сдохли!!!)\nВаши характеристики не поменялись, но зато да.",
								tWomb.ID, bot)
							time.Sleep(5 * time.Second)
							break
						} else if int(h2)-int(f1) <= 5 {
							editMsg(war1, fmt.Sprintf(
								"В раунде %d благодаря своей силе победил вомбат...",
								round), peer, bot)
							editMsg(war2, fmt.Sprintf(
								"В раунде %d благодаря лишению у другого здоровья победил вомбат...",
								round), tWomb.ID, bot)
							time.Sleep(3 * time.Second)
							h1c := int(womb.Health) / ((5 + rand.Intn(5)) / (rand.Intn(1) + 1))
							f1c := int(womb.Force) / ((5 + rand.Intn(5)) / (rand.Intn(1) + 1))
							mc := int((rand.Intn(int(womb.Health)) + 1) / 2)
							womb.Health += uint32(h1c)
							womb.Force += uint32(f1c)
							womb.Money += uint64(mc)
							womb.XP += 10
							editMsg(war1, fmt.Sprintf(
								"Победил вомбат %s!!!\nВы получили 10 XP, %d силы, %d здоровья и %d шишей, теперь их у Вас %d, %d, %d и %d соответственно",
								womb.Name, h1c, f1c, mc, womb.XP, womb.Health, womb.Force, womb.Money), peer, bot)
							tWomb.Health = 5
							tWomb.Money = 50
							editMsg(war2, fmt.Sprintf(
								"Победил вомбат %s!!!\nВаше здоровье обнулилось, а ещё у вас теперь только 50 шишей при себе :(",
								womb.Name), tWomb.ID, bot)
							break
						} else if int(h1)-int(f2) <= 5 {
							editMsg(war1, fmt.Sprintf(
								"В раунде %d благодаря своей силе победил вомбат...",
								round), peer, bot)
							editMsg(war2, fmt.Sprintf(
								"В раунде %d благодаря лишению у другого здоровья победил вомбат...",
								round), tWomb.ID, bot)
							time.Sleep(3 * time.Second)
							h2c := int(tWomb.Health) / ((5 + rand.Intn(5)) / (rand.Intn(1) + 1))
							f2c := int(tWomb.Force) / ((5 + rand.Intn(5)) / (rand.Intn(1) + 1))
							mc := int((rand.Intn(int(tWomb.Health)) + 1) / 2)
							tWomb.Health += uint32(h2c)
							tWomb.Force += uint32(f2c)
							tWomb.Money += uint64(mc)
							tWomb.XP += 10
							editMsg(war2, fmt.Sprintf(
								"Победил вомбат %s!!!\nВы получили 10 XP, %d силы, %d здоровья и %d шишей, теперь их у Вас %d, %d, %d и %d соответственно",
								tWomb.Name, h2c, f2c, mc, tWomb.XP, tWomb.Health, tWomb.Force, tWomb.Money), tWomb.ID, bot)
							womb.Health = 5
							womb.Money = 50
							editMsg(war1, fmt.Sprintf(
								"Победил вомбат %s!!!\nВаше здоровье сбросилось до 5, а ещё у вас теперь только 50 шишей при себе :(",
								tWomb.Name), peer, bot)
							break
						} else if round == 3 {
							if h1 < h2 {
								h2c := int(tWomb.Health) / ((5 + rand.Intn(5)) / (rand.Intn(1) + 1))
								f2c := int(tWomb.Force) / ((5 + rand.Intn(5)) / (rand.Intn(1) + 1))
								mc := int((rand.Intn(int(tWomb.Health)) + 1) / 2)
								tWomb.Health += uint32(h2c)
								tWomb.Force += uint32(f2c)
								tWomb.Money += uint64(mc)
								tWomb.XP += 10
								editMsg(war2, fmt.Sprintf(
									"И победил вомбат %s на раунде %d!!!\nВы получили 10 XP, %d силы, %d здоровья и %d шишей, теперь их у Вас %d, %d, %d и %d соответственно",
									tWomb.Name, round, h2c, f2c, mc, tWomb.XP, tWomb.Health, tWomb.Force, tWomb.Money), tWomb.ID, bot)
								womb.Health = uint32(h1)
								womb.Money = 50
								editMsg(war1, fmt.Sprintf(
									"И победил вомбат %s на раунде %d!\n К сожалению, теперь у вас только %d здоровья и 50 шишей при себе :(",
									tWomb.Name, round, womb.Health), peer, bot)
							} else {
								h1c := int(womb.Health) / ((5 + rand.Intn(5)) / (rand.Intn(1) + 1))
								f1c := int(womb.Force) / ((5 + rand.Intn(5)) / (rand.Intn(1) + 1))
								mc := int((rand.Intn(int(womb.Health)) + 1) / 2)
								womb.Health += uint32(h1c)
								womb.Force += uint32(f1c)
								womb.Money += uint64(mc)
								womb.XP += 10
								editMsg(war1, fmt.Sprintf(
									"Победил вомбат %s!!!\nВы получили 10 XP, %d силы, %d здоровья и %d шишей, теперь их у Вас %d, %d, %d и %d соответственно",
									womb.Name, h1c, f1c, mc, womb.XP, womb.Health, womb.Force, womb.Money), peer, bot)
								tWomb.Health = 5
								tWomb.Money = 50
								editMsg(war2, fmt.Sprintf(
									"Победил вомбат %s!!!\nВаше здоровье обнулилось, а ещё у вас теперь только 50 шишей при себе :(",
									womb.Name), tWomb.ID, bot)
							}
						}
					}
					err = docUpd(womb, wFil, users)
					if err != nil {
						replyToMsg(messID, errStart+"attack: accept: update_to", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
					docUpd(tWomb, bson.M{"_id": tWomb.ID}, users)
					if err != nil {
						replyToMsg(messID, errStart+"attack: accept: update_from", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
					_, err = attacks.DeleteOne(ctx, bson.M{"_id": at.ID})
					if err != nil {
						replyToMsg(messID, errStart+"attack: accept: delete", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
				default:
					replyToMsg(messID, "не понимаю!", peer, bot)
				}
			} else if isInList(txt, []string{"лечь спать", "споке", "спать", "споть"}) {
				if !isInUsers {
					sendMsg("У тебя нет вомбата, иди спи сам", peer, bot)
					return
				} else if womb.Sleep {
					sendMsg("Твой вомбат уже спит. Если хочешь проснуться, то напиши `проснуться` (логика)", peer, bot)
					return
				}
				womb.Sleep = true
				err = docUpd(womb, wFil, users)
				if err != nil {
					replyToMsg(messID, errStart+"go_sleep: update", peer, bot)
					rlog.Println("Error: ", err)
					return
				}
				sleep, err := getImgs(imgsC, "sleep")
				if err != nil {
					replyToMsg(messID, errStart+"go_sleep: get_imgs", peer, bot)
					rlog.Println("Error: ", err)
					return
				}
				sendPhoto(randImg(sleep), "Вы легли спать. Спокойного сна!", peer, bot)
			} else if isInList(txt, []string{"добрутро", "проснуться", "не спать", "не споть", "рота подъём"}) {
				if !isInUsers {
					sendMsg("У тебя нет вомбата, буди себя сам", peer, bot)
					return
				} else if !womb.Sleep {
					sendMsg("Твой вомбат и так не спит, может ты хотел лечь спать? (команда `лечь спать` (опять логика))",
						peer, bot)
					return
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
				err := docUpd(womb, wFil, users)
				if err != nil {
					replyToMsg(messID, errStart+"unsleep: update", peer, bot)
					rlog.Println("Error: ", err)
					return
				}
				unsleep, err := getImgs(imgsC, "unsleep")
				if err != nil {
					replyToMsg(messID, errStart+"unsleep: get_imgs", peer, bot)
					rlog.Println("Error: ", err)
					return
				}
				sendPhoto(randImg(unsleep), msg, peer, bot)
			} else if isPrefixInList(txt, []string{"рейтинг", "топ"}) {
				args := strings.Fields(strings.ToLower(txt))
				if args[0] != "рейтинг" && args[0] != "топ" {
					return
				}
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
						replyToMsg(messID, fmt.Sprintf("не понимаю, что значит %s", args[1]), peer, bot)
						return
					}
					if len(args) == 3 {
						if isInList(args[2], []string{"+", "плюс", "++", "увеличение"}) {
							queue = 1
						} else if isInList(args[2], []string{"-", "минус", "--", "уменьшение"}) {
							queue = -1
						} else {
							replyToMsg(messID, fmt.Sprintf("не понимаю, что значит %s", args[2]), peer, bot)
							return
						}
					}
				} else if len(args) != 1 {
					replyToMsg(messID, "слишком много аргументов", peer, bot)
					return
				}
				opts := options.Find()
				opts.SetSort(bson.M{name: queue})
				opts.SetLimit(10)
				cur, err := users.Find(ctx, bson.M{}, opts)
				if err != nil {
					replyToMsg(messID, errStart+"rating: find", peer, bot)
					rlog.Println("Error: ", err)
					return
				}
				var rating []User
				for cur.Next(ctx) {
					var w User
					cur.Decode(&w)
					rating = append(rating, w)
				}
				var msg string = fmt.Sprintf("Топ-10 вомбатов по ")
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
					msg += "увеличения"
				} else if queue == -1 {
					msg += "уменьшения"
				} else {
					replyToMsg(messID, errStart+"rating: queue else", peer, bot)
					rlog.Println("Error: rating: queue else")
					return
				}
				msg += ":\n"
				for num, w := range rating {
					switch name {
					case "money":
						msg += fmt.Sprintf("%d | %s (ID: %d) | %d шишей при себе\n", num+1, w.Name, w.ID, w.Money)
					case "xp":
						msg += fmt.Sprintf("%d | %s (ID: %d) | %d XP\n", num+1, w.Name, w.ID, w.XP)
					case "health":
						msg += fmt.Sprintf("%d | %s (ID: %d) | %d здоровья\n", num+1, w.Name, w.ID, w.Health)
					case "force":
						msg += fmt.Sprintf("%d | %s (ID: %d) | %d мощи\n", num+1, w.Name, w.ID, w.Force)
					}
				}
				msg = strings.TrimSuffix(msg, "\n")
				sendMsg(msg, peer, bot)
			} else if strings.HasPrefix(txt, "sendimg") {
				id := strings.TrimSpace(strings.TrimPrefix(txt, "sendimg"))
				sendPhoto(id, "", peer, bot)
			} else if strings.HasPrefix(strings.ToLower(txt), "вомбанк") {
				args := strings.Fields(strings.ToLower(txt))
				if len(args) == 0 {
					replyToMsg(messID, "как", peer, bot)
					return
				} else if args[0] != "вомбанк" {
					return
				}
				rCount, err := bank.CountDocuments(ctx, wFil)
				if err != nil {
					replyToMsg(messID, errStart+"bank: isBanked_count", peer, bot)
					rlog.Println("Error: ", err)
					return
				}
				isBanked := rCount == 1
				if len(args) == 0 {
					return
				} else if len(args) == 1 {
					replyToMsg(messID, "вомбанк", peer, bot)
				}
				switch args[1] {
				case "начать":
					if len(args) != 2 {
						replyToMsg(messID, "Вомбанк начать: слишком много аргументов", peer, bot)
						return
					} else if isBanked {
						replyToMsg(messID, "Ты уже зарегестрирован в вомбанке...", peer, bot)
						return
					} else if !isInUsers {
						replyToMsg(messID, "Вомбанк вомбатам! У тебя нет вомбата", peer, bot)
						return
					}
					b := Banked{
						ID:    from,
						Money: 15,
					}
					_, err = bank.InsertOne(ctx, b)
					if err != nil {
						replyToMsg(messID, errStart+"bank: new: insert", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
					replyToMsg(messID, "Вы были зарегестрированы в вомбанке! Вам на вомбосчёт добавили бесплатные 15 шишей",
						peer, bot)
				case "статус":
					var (
						fil   bson.M
						tWomb User
					)
					switch len(args) {
					case 2:
						if !isInUsers {
							replyToMsg(messID, "Вомбанк вомбатам! У тебя нет вомбата", peer, bot)
							return
						} else if !isBanked {
							replyToMsg(messID, "Вы не можете посмотреть вомбосчёт, которого нет", peer, bot)
							return
						}
						fil = wFil
						tWomb = womb
					case 3:
						if id, err := strconv.Atoi(args[2]); err == nil {
							fil = bson.M{"_id": id}
							rCount, err := users.CountDocuments(ctx, fil)
							if err != nil {
								replyToMsg(messID, errStart+"bank: status: count_user", peer, bot)
								rlog.Println("Error: ", err)
								return
							}
							if rCount != 1 {
								replyToMsg(messID, fmt.Sprintf("Вомбанк статус: пользователя с ID %d не найдено", id), peer, bot)
								return
							}
							bCount, err := bank.CountDocuments(ctx, fil)
							if err != nil {
								replyToMsg(messID, errStart+"bank: status: count_banked", peer, bot)
								rlog.Println("Error: ", err)
								return
							}
							if bCount != 1 {
								replyToMsg(messID, fmt.Sprintf("Вомбанк статус: у пользователя с ID %d нет вомбосчёта", id), peer, bot)
								return
							}
							err = users.FindOne(ctx, fil).Decode(&tWomb)
							if err != nil {
								replyToMsg(messID, errStart+"bank: status: find_womb", peer, bot)
								rlog.Println("Error: ", err)
								return
							}
						} else if id, ok := womb.Subs[args[2]]; ok {
							fil = bson.M{"_id": id}
							rCount, err := users.CountDocuments(ctx, fil)
							if err != nil {
								replyToMsg(messID, errStart+"bank: status: count_user", peer, bot)
								rlog.Println("Error: ", err)
								return
							}
							if rCount != 1 {
								replyToMsg(messID, fmt.Sprintf("Вомбанк статус: пользователя с ID %d не найдено", id), peer, bot)
								return
							}
							bCount, err := bank.CountDocuments(ctx, fil)
							if err != nil {
								replyToMsg(messID, errStart+"bank: status: count_banked", peer, bot)
								rlog.Println("Error: ", err)
								return
							}
							if bCount != 1 {
								replyToMsg(messID, fmt.Sprintf("Вомбанк статус: у пользователя %s нет вомбосчёта", args[2]), peer, bot)
								return
							}
							err = users.FindOne(ctx, fil).Decode(&tWomb)
							if err != nil {
								replyToMsg(messID, errStart+"bank: status: find_womb", peer, bot)
								rlog.Println("Error: ", err)
								return
							}
						} else {
							if len([]rune(args[2])) > 64 {
								replyToMsg(messID, "Слишком длинный алиас...", peer, bot)
								return
							}
							replyToMsg(messID, fmt.Sprintf("Вомбанк статус: подписчика с алиасом `%s` не найдено", args[2]), peer, bot)
							return
						}
					default:
						replyToMsg(messID, "Вомбанк статус: слишком много аргументов", peer, bot)
					}
					var b Banked
					err = bank.FindOne(ctx, fil).Decode(&b)
					if err != nil {
						replyToMsg(messID, errStart+"bank: status: find", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
					replyToMsg(messID, fmt.Sprintf(
						"Вомбанк вомбата %s (ID: %d):\nНа счету: %d\nПри себе: %d",
						tWomb.Name, tWomb.ID, b.Money, tWomb.Money), peer, bot)
				case "положить":
					if !isInUsers {
						replyToMsg(messID, "У тебя нет вомбата...", peer, bot)
						return
					} else if len(args) != 3 {
						replyToMsg(messID, "Вомбанк положить: недостаточно аргументов", peer, bot)
						return
					}
					if num, err := strconv.ParseUint(args[2], 10, 64); err == nil {
						if womb.Money < num+1 {
							replyToMsg(messID, "Вомбанк положить: недостаточно шишей при себе для операции", peer, bot)
							return
						} else if !isBanked {
							replyToMsg(messID,
								"Вомбанк положить: у вас нет ячейки в банке! Заведите её через `вомбанк начать`", peer, bot)
							return
						} else if num == 0 {
							replyToMsg(messID, "Ну и зачем?)", peer, bot)
							return
						}
						var b Banked
						err = bank.FindOne(ctx, wFil).Decode(&b)
						if err != nil {
							replyToMsg(messID, errStart+"bank: put: find_banked", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						womb.Money -= num
						b.Money += num
						err = docUpd(womb, wFil, users)
						if err != nil {
							replyToMsg(messID, errStart+"bank: put: upd_womb", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						err = docUpd(b, wFil, bank)
						if err != nil {
							replyToMsg(messID, errStart+"bank: put: upd_bank", peer, bot)
							rlog.Println("Error: ", err)
							return
						}
						replyToMsg(messID, fmt.Sprintf(
							"Ваш вомбосчёт пополнен на %d ш! Вомбосчёт: %d ш; При себе: %d ш",
							num, b.Money, womb.Money,
						), peer, bot)
					} else {
						replyToMsg(messID, "Вомбанк положить: требуется целое неотрицательное число шишей до 2^64", peer, bot)
					}
				case "снять":
					if !isInUsers {
						replyToMsg(messID, "У тебя нет вомбата...", peer, bot)
						return
					} else if len(args) != 3 {
						replyToMsg(messID, "Вомбанк снять: недостаточно аргументов", peer, bot)
						return
					}
					var b Banked
					err = bank.FindOne(ctx, wFil).Decode(&b)
					if err != nil {
						replyToMsg(messID, errStart+"bank: take: find_banked", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
					var num uint64
					var err error
					if num, err = strconv.ParseUint(args[2], 10, 64); err == nil {
						if num == 0 {
							replyToMsg(messID, "Ну и зачем?", peer, bot)
							return
						}
					} else if args[2] == "всё" || args[2] == "все" {
						if b.Money == 0 {
							replyToMsg(messID, "У вас на счету 0 шишей. Зачем?", peer, bot)
							return
						}
						num = b.Money
					} else {
						replyToMsg(messID, "Вомбанк снять: требуется целое неотрицательное число шишей до 2^64", peer, bot)
						return
					}
					if b.Money < num {
						replyToMsg(messID, "Вомбанк снять: недостаточно шишей на вомбосчету для операции", peer, bot)
						return
					} else if !isBanked {
						replyToMsg(messID,
							"Вомбанк снять: у вас нет ячейки в банке! Заведите её через `вомбанк начать`", peer, bot)
						return
					}
					b.Money -= num
					womb.Money += num
					err = docUpd(b, wFil, bank)
					if err != nil {
						replyToMsg(messID, errStart+"bank: put: upd_bank", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
					err = docUpd(womb, wFil, users)
					if err != nil {
						replyToMsg(messID, errStart+"bank: put: upd_womb", peer, bot)
						rlog.Println("Error: ", err)
						return
					}
					replyToMsg(messID, fmt.Sprintf(
						"Вы сняли %d ш! Вомбосчёт: %d ш; При себе: %d ш",
						num, b.Money, womb.Money,
					), peer, bot)
				case "перевести":
					if len(args) != 5 {
						replyToMsg(messID, "Вомбанк перевести: слишком мало или много аргументов", peer, bot)
						return
					}
				default:
					replyToMsg(messID, "Вомбанк: неизвестная команда: "+args[1], peer, bot)
				}
			}
		}(update, titles, bot, users, titlesC, attacks, imgsC, bank)
	}
}
