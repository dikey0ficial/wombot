package main

import (
	"context"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/caarlos0/env"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

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
	Sleep  bool     `bson:"sleep"`
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

// Clan реализует клан
type Clan struct {
	Tag            string    `bson:"_id"`
	Name           string    `bson:"name"`
	Money          uint64    `bson:"money"` // Казна
	XP             uint32    `bson:"xp"`
	Leader         int64     `bson:"leader"`
	Banker         int64     `bson:"banker"`
	Members        []int64   `bson:"members"`
	Banned         []int64   `bson:"banned"`
	LastRewarsTime time.Time `bson:"last_reward_time"`
}

// Clattack реализует клановую атаку
type Clattack struct {
	ID   string `bson:"_id"`
	From string `bson:"from"`
	To   string `bson:"to"`
}

// Clwar реализует клана-бойца
type Clwar struct {
	Tag    string
	Name   string
	Health uint32
	Force  uint32
}

var (
	infl, errl, debl, servl *log.Logger
	ctx                     = context.Background()
	conf                    = struct {
		Token     string `toml:"tg_token" env:"TGTOKEN"`
		MongoURL  string `toml:"mongo_url" env:"MONGOURL"`
		SupChatID int64  `toml:"support_chat_id" env:"SUPCHATID"`
	}{}
)

func init() {
	infl = log.New(os.Stdout, "[INFO]\t", log.Ldate|log.Ltime)
	errl = log.New(os.Stderr, "[ERROR]\t", log.Ldate|log.Ltime|log.Lshortfile)
	debl = log.New(os.Stdout, "[DEBUG]\t", log.Ldate|log.Ltime|log.Lshortfile)
	if f, err := os.Open("config.toml"); err == nil {
		dat, err := ioutil.ReadAll(f)
		if err != nil {
			errl.Println(err)
			os.Exit(1)
		}
		if err := toml.Unmarshal(dat, &conf); err != nil {
			errl.Println(err)
			os.Exit(1)
		}
	} else if !os.IsNotExist(err) {
		if err := env.Parse(&conf); err != nil {
			errl.Println(err)
			os.Exit(1)
		}
	} else {
		errl.Println(err)
		os.Exit(1)
	}
}

// checkerr реализует проверку ошибок без паники
func checkerr(err error) {
	if err != nil && err.Error() != "EOF" {
		errl.Printf("e: %v\n", err)
	}
}

// isInList нужен для проверки сообщений
func isInList(str string, list []string) bool {
	for _, elem := range list {
		if strings.ToLower(str) == strings.ToLower(elem) {
			return true
		}
	}
	return false
}

// hasTitle _
func hasTitle(i uint16, list []uint16) bool {
	for _, elem := range list {
		if i == elem {
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
	if err != nil {
		log.Println(chatID)
	}
	return mess.MessageID
}

// sendMsgMD отправляет сообщение с markdown
func sendMsgMD(message string, chatID int64, bot *tg.BotAPI) int {
	msg := tg.NewMessage(chatID, message)
	mess, err := bot.Send(msg)
	msg.ParseMode = "markdown"
	checkerr(err)
	if err != nil {
		log.Println(chatID)
	}
	return mess.MessageID
}

// replyToMsg отвечает обычным сообщением
func replyToMsg(replyID int, message string, chatID int64, bot *tg.BotAPI) int {
	msg := tg.NewMessage(chatID, message)
	msg.ReplyToMessageID = replyID
	mess, err := bot.Send(msg)
	checkerr(err)
	if err != nil {
		log.Println(chatID)
	}
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
	if err != nil {
		log.Println(chatID)
	}
	return mess.MessageID
}

// sendPhoto отправляет текст с картинкой
func sendPhoto(id, caption string, chatID int64, bot *tg.BotAPI) int {
	msg := tg.NewPhoto(chatID, tg.FileID(id))
	msg.Caption = caption
	mess, err := bot.Send(msg)
	checkerr(err)
	if err != nil {
		log.Println(chatID)
	}
	return mess.MessageID
}

// sendPhotoMD отправляет текст с markdown с картинкой
func sendPhotoMD(id, caption string, chatID int64, bot *tg.BotAPI) int {
	msg := tg.NewPhoto(chatID, tg.FileID(id))
	msg.Caption = caption
	msg.ParseMode = "markdown"
	mess, err := bot.Send(msg)
	checkerr(err)
	if err != nil {
		log.Println(chatID)
	}
	return mess.MessageID
}

// replyToMsgMD отвечает сообщением с markdown
func replyToMsgMD(replyID int, message string, chatID int64, bot *tg.BotAPI) int {
	msg := tg.NewMessage(chatID, message)
	msg.ReplyToMessageID = replyID
	msg.ParseMode = "markdown"
	mess, err := bot.Send(msg)
	checkerr(err)
	if err != nil {
		log.Println(chatID)
	}
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
	if err != nil {
		log.Println(chatID)
	}
	return mess.MessageID
}

// replyWithPhotoMD отвечает картинкой с текстом
func replyWithPhoto(replyID int, id, caption string, chatID int64, bot *tg.BotAPI) int {
	msg := tg.NewPhoto(chatID, tg.FileID(id))
	msg.Caption = caption
	msg.ReplyToMessageID = replyID
	mess, err := bot.Send(msg)
	checkerr(err)
	if err != nil {
		log.Println(chatID)
	}
	return mess.MessageID
}

// isInAttacks возвращает информацию, есть ли существо в атаках и
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

// isInClttacks возвращает информацию, есть ли клан в клановых атаках и
// отправитель ли он
func isInClattacks(id string, attacks *mongo.Collection) (isIn, isFrom bool) {
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

func caseInsensitive(s string) primitive.Regex {
	return primitive.Regex{
		Pattern: fmt.Sprintf("^%s$", s),
		Options: "i",
	}
}

var (
	valids []string = []string{
		"qwertyuiopasdfghjklzxcvbnm",
		"QWERTYUIOPASDFGHJKLZXCVBNM",
		"ёйцукенгшщзхъфывапролджэячсмитьбю",
		"ЁЙЦУКЕНГШЩЗХЪФЫВАПРОЛДЖЭЯЧСМИТЬБЮ",
		"0123456789",
		"_-:()~ε",
	}
	validName string = strings.Join(valids, "")
)

func isValidName(name string) bool {
	for _, nl := range name {
		is := false
		for _, sym := range validName {
			if nl == sym {
				is = true
				break
			}
		}
		if !is {
			return is
		}
	}
	return true
}

func isValidTag(tag string) bool {
	for _, nl := range strings.ToLower(tag) {
		is := false
		for _, sym := range valids[0] {
			if nl == sym {
				is = true
				break
			}
		}
		if !is {
			return is
		}
	}
	return true
}

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

	clans := db.Collection("clans")
	clattacks := db.Collection("clattacks")

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
	infl.Println("Started bot")

	for update := range updates {
		if update.Message == nil {
			continue
		} else if update.Message.Photo != nil {
			infl.Println("img ", (update.Message.Photo)[0].FileID)
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
					errl.Println("e: ", err)
					return
				}
				isInUsers := rCount != 0
				if isInUsers {
					err = users.FindOne(ctx, wFil).Decode(&womb)
					if err != nil {
						replyToMsg(messID, errStart+"womb: find_womb", peer, bot)
						errl.Println("e: ", err)
						return
					}
				}

				if update.Message.NewChatMembers != nil && len(update.Message.NewChatMembers) != 0 {
					if !isInUsers {
						replyToMsgMDNL(messID,
							"Здравствуйте! Я [вомбот](t.me/wombatobot) — бот с вомбатами. "+
								"Рекомендую Вам завести вомбата, чтобы играть "+
								"вместе с другими участниками этого чата (^.^)",
							peer, bot,
						)
					} else {
						replyToMsgMD(messID, fmt.Sprintf("Добро пожаловать, вомбат `%s`!", womb.Name), peer, bot)
					}
					return
				}

				infl.Printf("[GROUP_MESSGAE] i:%d p:%d f:%d un:%s, wn:%s, t:%s\n", messID, peer, from,
					update.Message.From.UserName, womb.Name,
					strings.Join(strings.Fields(txt), " "))
				if strings.HasPrefix(strings.ToLower(txt), "о вомбате") {
					strID := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(txt), "о вомбате"))
					var (
						tWomb User
					)
					if strID == "" {
						if isInUsers {
							tWomb = womb
						} else {
							replyToMsg(messID, "У вас нет вомбата", peer, bot)
							return
						}
					} else if len([]rune(strID)) > 64 {
						replyToMsg(messID, "Ошибка: слишком длинное имя", peer, bot)
						return
					} else if !isValidName(strID) {
						replyToMsg(messID, "Нелегальное имя!", peer, bot)
						return
					} else if rCount, err :=
						users.CountDocuments(ctx, bson.M{"name": caseInsensitive(strID)}); err == nil && rCount != 0 {
						err := users.FindOne(ctx, bson.M{"name": caseInsensitive(strID)}).Decode(&tWomb)
						if err != nil {
							replyToMsg(messID, errStart+"about_womb: find_users_name", peer, bot)
							errl.Println("e: ", err)
							return
						}
					} else if err != nil {
						replyToMsg(messID, errStart+"about_womb: count_users_name", peer, bot)
						errl.Println("e: ", err)
						return
					} else {
						replyToMsg(messID, fmt.Sprintf("Ошибка: пользователя с именем %s не найдено", strID), peer, bot)
						return
					}
					strTitles := ""
					tCount := len(tWomb.Titles)
					if tCount != 0 {
						for _, id := range tWomb.Titles {
							rCount, err = titlesC.CountDocuments(ctx, bson.M{"_id": id})
							if err != nil {
								replyToMsg(messID, errStart+"about_womb: count_titles", peer, bot)
								errl.Println("e: ", err)
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
								errl.Println("e: ", err)
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
					abimg, err := getImgs(imgsC, "about")
					if err != nil {
						replyToMsg(messID, errStart+"about_womb: get_imgs", peer, bot)
						errl.Println("e: ", err)
						return
					}
					replyWithPhotoMD(messID, randImg(abimg), fmt.Sprintf(
						"Вомбат `%s`\nТитулы: %s\n 👁 %d XP\n ❤ %d здоровья\n ⚡ %d мощи\n 💰 %d шишей при себе\n 💤 %s",
						tWomb.Name, strTitles, tWomb.XP, tWomb.Health, tWomb.Force, tWomb.Money, sl),
						peer, bot,
					)
				} else if strings.HasPrefix(strings.ToLower(txt), "хрю") {
					hru, err := getImgs(imgsC, "schweine")
					if err != nil {
						replyToMsg(messID, errStart+"schweine: get_imgs", peer, bot)
						errl.Println("e: ", err)
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
							errl.Println("e: ", err)
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
							if len([]rune(strID)) > 64 {
								replyToMsg(messID, "Слишком длинное имя!", peer, bot)
								return
							} else if !isValidName(strID) {
								replyToMsg(messID, "нелегальный никнейм!", peer, bot)
								return
							} else if rCount, err := users.CountDocuments(ctx,
								bson.M{"name": caseInsensitive(strID)}); err == nil && rCount != 0 {
								var tWomb User
								err = users.FindOne(ctx, bson.M{"name": caseInsensitive(strID)}).Decode(&tWomb)
								if err != nil {
									replyToMsg(messID, errStart+"attack: find_users_name", peer, bot)
									errl.Println("e: ", err)
									return
								}
								ID = tWomb.ID
							} else if err != nil {
								replyToMsg(messID, errStart+"attack: status: count_users_name", peer, bot)
								errl.Println("e: ", err)
								return
							} else {
								replyToMsg(messID, fmt.Sprintf("Пользователя с никнеймом `%s` не найдено", strID), peer, bot)
								return
							}
						}
						var at Attack
						if is, isFrom := isInAttacks(ID, attacks); isFrom {
							a, err := getAttackByWomb(ID, true, attacks)
							if err != nil {
								replyToMsg(messID, errStart+"attack: status: to_at", peer, bot)
								errl.Println("e: ", err)
								return
							}
							at = a
						} else if is {
							a, err := getAttackByWomb(from, false, attacks)
							if err != nil {
								replyToMsg(messID, errStart+"attack: status: from_at", peer, bot)
								errl.Println("e: ", err)
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
							errl.Println("e: ", err)
							return
						}
						err = users.FindOne(ctx, bson.M{"_id": at.To}).Decode(&toWomb)
						if err != nil {
							replyToMsg(messID, errStart+"attack: status: find_twomb", peer, bot)
							errl.Println("e: ", err)
							return
						}
						replyToMsg(messID, fmt.Sprintf(
							"От: %s\nКому: %s\n",
							fromWomb.Name, toWomb.Name,
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
						errl.Println("e: ", err)
						return
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
					default:
						replyToMsg(messID, errStart+"rating: name else", peer, bot)
						errl.Println("e: rating: name else")
						return
					}
					msg += "в порядке "
					if queue == 1 {
						msg += "увеличения:"
					} else if queue == -1 {
						msg += "уменьшения:"
					} else {
						replyToMsg(messID, errStart+"rating: queue else", peer, bot)
						errl.Println("RROR err:rating: queue else")
						return
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
					replyToMsg(messID, msg, peer, bot)
				}
			}(update, titles, bot, users, titlesC, attacks, imgsC, bank)
			continue
		}
		// MESSAGE_DIRECT_CHAT
		go func(update tg.Update, titles []Title, bot *tg.BotAPI, users, titlesC,
			attacks, imgsC, bank, clans, clattacks *mongo.Collection) {

			peer, from := update.Message.Chat.ID, update.Message.From.ID
			txt, messID := strings.TrimSpace(update.Message.Text), update.Message.MessageID

			const errStart string = "Ошибка... Ответьте командой /admin на это сообщение\n"

			womb := User{}

			wFil := bson.M{"_id": from}

			rCount, err := users.CountDocuments(ctx, wFil)
			if err != nil {
				replyToMsg(messID, errStart+"isInUsers: count_womb", peer, bot)
				errl.Println("e: ", err)
				return
			}
			isInUsers := rCount != 0
			if isInUsers {
				err = users.FindOne(ctx, wFil).Decode(&womb)
				if err != nil {
					replyToMsg(messID, errStart+"womb: find_womb", peer, bot)
					errl.Println("e: ", err)
					return
				}
			}

			infl.Printf("[DIRECT_MESSAGE] i:%d, p:%d f:%d un:%s wn:%s t:%s\n", messID, peer, from,
				update.Message.From.UserName, womb.Name,
				strings.Join(strings.Fields(txt), " "))

			if strings.HasPrefix(txt, "/start") {
				sendMsg("Привет! \n — Завести вомбата: `взять вомбата`\n — Помощь: https://telegra.ph/Pomoshch-10-28 (/help)\n — Канал бота, где есть нужная инфа: @wombatobot_channel\n Приятной игры!",
					peer, bot,
				)
			} else if isInList(txt, []string{"взять вомбата", "купить вомбата у арабов", "хочу вомбата"}) {
				if isInUsers {
					replyToMsg(messID,
						"У тебя как бы уже есть вомбат лолкек. Если хочешь от него избавиться, то напиши `приготовить шашлык`",
						peer, bot,
					)
				} else {
					rand.Seed(peer)
					newWomb := User{ID: peer,
						Name:   "Вомбат_" + strconv.Itoa(int(from)),
						XP:     0,
						Health: 5,
						Force:  2,
						Money:  10,
						Titles: []uint16{},
						Sleep:  false,
					}
					_, err = users.InsertOne(ctx, &newWomb)
					if err != nil {
						replyToMsg(messID, errStart+"new_womb: insert", peer, bot)
						errl.Println("e: ", err)
						return
					}
					newimg, err := getImgs(imgsC, "new")
					if err != nil {
						replyToMsg(messID, errStart+"new_womb: get_imgs", peer, bot)
						errl.Println("e: ", err)
						return
					}
					replyWithPhoto(messID,
						randImg(newimg), fmt.Sprintf(
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
								errl.Println("e: ", err)
								return
							}
							replyToMsg(messID, fmt.Sprintf("Операция проведена успешно! Шишей при себе: %d", womb.Money), peer, bot)
						} else {
							replyToMsg(messID,
								"Ошибка: неправильный синтаксис. Синтаксис команды: `devtools set money {кол-во шишей}`", peer, bot)
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
							replyToMsg(messID,
								"Ошибка: неправильный синтаксис. Синтаксис команды: `devtools reset [force/health/xp/all]`",
								peer, bot,
							)
							return
						}
						err := docUpd(womb, wFil, users)
						if err != nil {
							replyToMsg(messID, errStart+"devtools: reset: update", peer, bot)
							errl.Println("e: ", err)
							return
						}
						replyToMsg(messID, "Операция произведена успешно!", peer, bot)
					} else if cmd == "help" {
						replyToMsg(messID, "https://telegra.ph/Vombot-devtools-help-10-28", peer, bot)
					}
				} else if strings.TrimSpace(txt) == "devtools on" {
					womb.Titles = append(womb.Titles, 0)
					err := docUpd(womb, wFil, users)
					if err != nil {
						replyToMsg(messID, errStart+"devtools: on", peer, bot)
						errl.Println("e: ", err)
						return
					}
					replyToMsg(messID, "Выдан титул \"Вомботестер\" (ID: 0)", peer, bot)
				}
			} else if isInList(txt, []string{"приготовить шашлык", "продать вомбата арабам", "слить вомбата в унитаз", "убить"}) {
				if isInUsers {
					if !(hasTitle(1, womb.Titles)) {
						_, err = users.DeleteOne(ctx, wFil)
						if err != nil {
							replyToMsg(messID, errStart+"kill: delete", peer, bot)
							errl.Println("e: ", err)
							return
						}
						kill, err := getImgs(imgsC, "kill")
						if err != nil {
							replyToMsg(messID, errStart+"kill: get_imgs", peer, bot)
							errl.Println("e: ", err)
							return
						}
						replyWithPhoto(messID,
							randImg(kill), "Вы уничтожили вомбата в количестве 1 штука. Вы - нехорошее существо", peer, bot)
					} else {
						replyToMsg(messID,
							"Ошибка: вы лишены права уничтожать вомбата; ответьте на это сообщение командой /admin для объяснений",
							peer, bot)
					}
				} else {
					replyToMsg(messID, "Но у вас нет вомбата...", peer, bot)
				}
			} else if args := strings.Fields(txt); len(args) > 1 && strings.ToLower(strings.Join(args[:2], " ")) == "поменять имя" {
				if !isInUsers {
					replyToMsg(messID, "Да блин нафиг, вы вобмата забыли завести!!!!!!!", peer, bot)
				} else if len(args) != 3 {
					if len(args) == 2 {
						replyToMsg(messID, "вомбату нужно имя! ты его не указал", peer, bot)
					} else {
						replyToMsg(messID, "слишком много аргументов...", peer, bot)
					}
					return
				} else if hasTitle(1, womb.Titles) {
					replyToMsg(messID, "Тебе нельзя, ты спамер (оспорить: /admin)", peer, bot)
					return
				} else if womb.Money < 3 {
					replyToMsg(messID, "Мало шишей блин нафиг!!!!", peer, bot)
					return
				}
				name := args[2]
				if womb.Name == name {
					replyToMsg(messID, "зачем", peer, bot)
					return
				} else if len([]rune(name)) > 64 {
					replyToMsg(messID, "Слишком длинный никнейм!", peer, bot)
					return
				} else if isInList(name, []string{"вoмбoт", "вoмбoт", "вомбoт", "вомбот", "бот", "bot", "бoт", "bоt",
					"авто", "auto"}) {
					replyToMsg(messID, "Такие никнеймы заводить нельзя", peer, bot)
				} else if !isValidName(name) {
					replyToMsg(messID, "Нелегальное имя:(\n", peer, bot)
					return
				}
				rCount, err := users.CountDocuments(ctx, bson.M{"name": caseInsensitive(name)})
				if err != nil {
					replyToMsg(messID, errStart+"rename: count", peer, bot)
					errl.Println("e: ", err)
					return
				} else if rCount != 0 {
					replyToMsg(messID, fmt.Sprintf("Никнейм `%s` уже занят(", name), peer, bot)
					return
				}
				womb.Money -= 3
				split := strings.Fields(txt)
				caseName := strings.Join(split[2:], " ")
				womb.Name = caseName
				err = docUpd(womb, wFil, users)
				if err != nil {
					replyToMsg(messID, errStart+"rename: update", peer, bot)
					errl.Println("e: ", err)
					return
				}
				replyToMsg(messID,
					fmt.Sprintf("Теперь вашего вомбата зовут %s. С вашего счёта сняли 3 шиша", caseName),
					peer, bot,
				)
			} else if isInList(txt, []string{"помощь", "хелп", "help", "команды", "/help", "/help@wombatobot"}) {
				replyToMsg(messID, "https://telegra.ph/Pomoshch-10-28", peer, bot)
			} else if strings.ToLower(txt) == "магазин" {
				replyToMsg(messID, strings.Join([]string{"Магазин:", " — 1 здоровье — 5 ш", " — 1 мощь — 3 ш",
					" — квес — 256 ш", " — вадшам — 250'000 ш",
					"Для покупки использовать команду 'купить [название_объекта] ([кол-во])",
				}, "\n"),
					peer, bot,
				)
			} else if args := strings.Fields(strings.ToLower(txt)); len(args) != 0 && args[0] == "купить" {
				if len(args) == 1 {
					replyToMsg(messID, "купить", peer, bot)
					return
				}
				switch args[1] {
				case "здоровья":
					fallthrough
				case "здоровье":
					if len(args) > 3 {
						replyToMsg(messID, "Ошибка: слишком много аргументов...", peer, bot)
						return
					}
					if isInUsers {
						var amount uint32 = 1
						if len(args) == 3 {
							if val, err := strconv.ParseUint(args[2], 10, 32); err == nil {
								if val == 0 {
									replyToMsg(messID, "Поздравляю! Теперь у вас одна шиза и ещё одна шиза", peer, bot)
									return
								}
								amount = uint32(val)
							} else {
								replyToMsg(messID, "Ошибка: число должно быть неотрицательным, целым и меньше 2^32", peer, bot)
								return
							}
						}
						if womb.Money >= uint64(amount*5) {
							if uint64(womb.Health+amount) < uint64(math.Pow(2, 32)) {
								womb.Money -= uint64(amount * 5)
								womb.Health += amount
								err := docUpd(womb, wFil, users)
								if err != nil {
									replyToMsg(messID, errStart+"buy: health: update", peer, bot)
									errl.Println("e: ", err)
									return
								}
								replyToMsg(messID,
									fmt.Sprintf("Поздравляю! Теперь у вас %d здоровья и %d шишей при себе", womb.Health, womb.Money),
									peer, bot)
							} else {
								replyToMsg(messID,
									"Ошибка: вы достигли максимального количества здоровья (2 в 32 степени). Если это вас возмущает, ответьте командой /admin",
									peer, bot,
								)
							}
						} else {
							replyToMsg(messID, "Надо накопить побольше шишей! 1 здоровье = 5 шишей", peer, bot)
						}
					} else {
						replyToMsg(messID, "У тя ваще вобата нет...", peer, bot)
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
						replyToMsg(messID, "Ошибка: слишком много аргументов...", peer, bot)
						return
					}
					if isInUsers {
						var amount uint32 = 1
						if len(args) == 3 {
							if val, err := strconv.ParseUint(args[2], 10, 32); err == nil {
								if val == 0 {
									replyToMsg(messID, "Поздравляю! Теперь у вас одна шиза и ещё одна шиза", peer, bot)
									return
								}
								amount = uint32(val)
							} else {
								replyToMsg(messID, "Ошибка: число должно быть неотрицательным, целым и меньше 2^32", peer, bot)
								return
							}
						}
						if womb.Money >= uint64(amount*3) {
							if uint64(womb.Force+1) < uint64(math.Pow(2, 32)) {
								womb.Money -= uint64(amount * 3)
								womb.Force += amount
								err := docUpd(womb, wFil, users)
								if err != nil {
									replyToMsg(messID, errStart+"buy: force: update", peer, bot)
									errl.Println("e: ", err)
									return
								}
								replyToMsg(messID, fmt.Sprintf("Поздравляю! Теперь у вас %d мощи и %d шишей при себе", womb.Force, womb.Money),
									peer, bot)
							} else {
								replyToMsg(messID,
									"Ошибка: вы достигли максимального количества мощи (2 в 32 степени). Если это вас возмущает, ответьте командой /admin",
									peer, bot,
								)
							}
						} else {
							replyToMsg(messID, "Надо накопить побольше шишей! 1 мощь = 3 шиша", peer, bot)
						}
					} else {
						replyToMsg(messID, "У тя ваще вобата нет...", peer, bot)
					}
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
						replyToMsg(messID, "ужас !! слишком много аргументов!!!", peer, bot)
						return
					} else if !isInUsers {
						replyToMsg(messID, "ты не можешь купить вадшарма без вомбата", peer, bot)
						return
					} else if hasTitle(4, womb.Titles) {
						replyToMsg(messID, "у вас уже есть вадшам", peer, bot)
						return
					} else if womb.Money < 250005 {
						replyToMsg(messID, "Ошибка: недостаточно шишей для покупки (требуется 250000 + 5)", peer, bot)
						return
					}
					womb.Money -= 250000
					womb.Titles = append(womb.Titles, 4)
					err = docUpd(womb, wFil, users)
					if err != nil {
						replyToMsg(messID, errStart+"buy: vadimushka: update", peer, bot)
						errl.Println("e: ", err)
						return
					}
					replyToMsg(messID, "Теперь вы вадшамообладатель", peer, bot)
				case "квес":
					fallthrough
				case "квеса":
					fallthrough
				case "квесу":
					fallthrough
				case "qwess":
					if len(args) != 2 {
						replyToMsg(messID, "Слишком много аргументов!", peer, bot)
						return
					} else if !isInUsers {
						replyToMsg(messID, "К сожалению, вам нужны шиши, чтобы купить квес, а шиши есть только у вомбатов...", peer, bot)
					} else if womb.Money < 256 {
						leps, err := getImgs(imgsC, "leps")
						if err != nil {
							replyToMsg(messID, errStart+"buy: nyamka: get_leps_imgs", peer, bot)
							errl.Println("e: ", err)
							return
						}
						replyWithPhoto(messID,
							randImg(leps),
							"Вы подошли к ближайшему ларьку, но, увы, кролик-Лепс на кассе сказал, что надо 256 шишей, а у вас, к сожалению, меньше",
							peer, bot,
						)
						return
					}
					qwess, err := getImgs(imgsC, "qwess")
					if err != nil {
						replyToMsg(messID, errStart+"buy: nyamka: get_qwess_imgs", peer, bot)
						errl.Println("e: ", err)
						return
					}
					if !(hasTitle(2, womb.Titles)) {
						womb.Titles = append(womb.Titles, 2)
						womb.Money -= 256
						err = docUpd(womb, wFil, users)
						if err != nil {
							replyToMsg(messID, errStart+"buy: nyamka: update_first_time", peer, bot)
							errl.Println("e: ", err)
							return
						}
						replyWithPhoto(messID,
							randImg(qwess),
							"Вы купили чудесного вкуса квес у кролика-Лепса в ларьке за 256 шишей. Глотнув этот напиток, вы поняли, что получили новый титул с ID 2",
							peer, bot,
						)
					} else {
						womb.Money -= 256
						err = docUpd(womb, wFil, users)
						if err != nil {
							replyToMsg(messID, errStart+"buy: nyamka: update", peer, bot)
							errl.Println("e: ", err)
							return
						}
						if err != nil {
							replyToMsg(messID, errStart+"buy: nyamka: get_leps_imgs", peer, bot)
							errl.Println("e: ", err)
							return
						}
						replyWithPhoto(messID,
							randImg(qwess),
							"Вы вновь купили вкусного квеса у того же кролика-Лепса в том же ларьке за 256 шишей. \"Он так освежает, я чувствую себя человеком\" — думаете вы. Ах, как вкусён квес!",
							peer, bot,
						)
					}
				default:
					replyToMsg(messID, fmt.Sprintf("Что такое %s?", args[1]), peer, bot)
				}
			} else if isInList(txt, []string{"поиск денег"}) {
				if isInUsers {
					if womb.Money < 5 {
						womb.Money = 5
						err := docUpd(womb, wFil, users)
						if err != nil {
							replyToMsg(messID, errStart+"find_money: free", peer, bot)
							errl.Println("e: ", err)
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
							errl.Println("e: ", err)
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
						errl.Println("e: ", err)
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
			} else if strings.HasPrefix(strings.ToLower(txt), "о вомбате") {
				strID := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(txt), "о вомбате"))
				var (
					tWomb User
				)
				if strID == "" {
					if isInUsers {
						tWomb = womb
					} else {
						replyToMsg(messID, "У вас нет вомбата", peer, bot)
						return
					}
				} else if len([]rune(strID)) > 64 {
					replyToMsg(messID, "Ошибка: слишком длинное имя", peer, bot)
					return
				} else if !isValidName(strID) {
					replyToMsg(messID, "Нелегальное имя", peer, bot)
					return
				} else if rCount, err :=
					users.CountDocuments(ctx, bson.M{"name": caseInsensitive(strID)}); err == nil && rCount != 0 {
					err := users.FindOne(ctx, bson.M{"name": caseInsensitive(strID)}).Decode(&tWomb)
					if err != nil {
						replyToMsg(messID, errStart+"about_womb: find_users_name", peer, bot)
						errl.Println("e: ", err)
						return
					}
				} else if err != nil {
					replyToMsg(messID, errStart+"about_womb: count_users_name", peer, bot)
					errl.Println("e: ", err)
					return
				} else {
					replyToMsg(messID, fmt.Sprintf("Ошибка: пользователя с именем %s не найдено", strID), peer, bot)
					return
				}
				strTitles := ""
				tCount := len(tWomb.Titles)
				if tCount != 0 {
					for _, id := range tWomb.Titles {
						rCount, err = titlesC.CountDocuments(ctx, bson.M{"_id": id})
						if err != nil {
							replyToMsg(messID, errStart+"about_womb: count_titles", peer, bot)
							errl.Println("e: ", err)
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
							errl.Println("e: ", err)
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
				}
				abimg, err := getImgs(imgsC, "about")
				if err != nil {
					replyToMsg(messID, errStart+"about_womb: get_imgs", peer, bot)
					errl.Println("e: ", err)
					return
				}
				replyWithPhotoMD(messID, randImg(abimg), fmt.Sprintf(
					"Вомбат `%s`\nТитулы: %s\n 🕳 %d XP\n ❤ %d здоровья\n ⚡ %d мощи\n 💰 %d шишей при себе\n 💤 %s",
					tWomb.Name, strTitles, tWomb.XP, tWomb.Health, tWomb.Force, tWomb.Money, sl),
					peer, bot,
				)
			} else if strings.HasPrefix(strings.ToLower(txt), "о вомботе") {
				sendMsgMD("https://telegra.ph/O-vombote-10-29\n**если вы хотели узнать характеристики вомбата, используйте команду `о вомбате`**",
					peer, bot,
				)
			} else if strings.HasPrefix(strings.ToLower(txt), "перевести шиши") {
				args := strings.Fields(strings.TrimSpace(strings.TrimPrefix(strings.ToLower(txt), "перевести шиши")))
				if len(args) < 2 {
					replyToMsg(messID,
						"Ошибка: вы пропустили аргумент(ы). Синтаксис команды: `перевести шиши [кол-во] [никнейм получателя]`",
						peer, bot,
					)
				} else if len(args) > 2 {
					replyToMsg(messID,
						"Ошибка: слишком много аргументов. Синтаксис команды: `перевести шиши [кол-во] [никнейм получателя]`",
						peer, bot,
					)
				} else {
					if amount, err := strconv.ParseUint(args[0], 10, 64); err == nil {
						var ID int64
						name := args[1]
						if len([]rune(name)) > 64 {
							replyToMsg(messID, "Слишком длинный никнейм", peer, bot)
							return
						} else if !isValidName(name) {
							replyToMsg(messID, "Нелегальное имя", peer, bot)
							return
						} else if rCount, err := users.CountDocuments(
							ctx, bson.M{"name": caseInsensitive(name)}); err == nil && rCount != 0 {
							var tWomb User
							err = users.FindOne(ctx, bson.M{"name": caseInsensitive(name)}).Decode(&tWomb)
							if err != nil {
								replyToMsg(messID, errStart+"send_shishs: status: find_users_name", peer, bot)
								errl.Println("e: ", err)
								return
							}
							ID = tWomb.ID
						} else if err != nil {
							replyToMsg(messID, errStart+"send_shishs: status: count_users_name", peer, bot)
							errl.Println("e: ", err)
							return
						} else {
							replyToMsg(messID, fmt.Sprintf("Ошибка: вомбата с именем %s не найдено", name), peer, bot)
							return
						}
						if womb.Money >= amount {
							if amount != 0 {
								if ID == peer {
									replyToMsg(messID, "Ты читер блин нафиг!!!!!! нидам тебе самому себе перевести", peer, bot)
									return
								}
								rCount, err = users.CountDocuments(ctx, bson.M{"_id": ID})
								if err != nil {
									replyToMsg(messID, errStart+"send_shishs: count_to", peer, bot)
									errl.Println("e: ", err)
									return
								}
								if rCount != 0 {
									tWomb := User{}
									err = users.FindOne(ctx, bson.M{"_id": ID}).Decode(&tWomb)
									if err != nil {
										replyToMsg(messID, errStart+"send_shishs: find_to", peer, bot)
										errl.Println("e: ", err)
										return
									}
									womb.Money -= amount
									tWomb.Money += amount
									err := docUpd(tWomb, bson.M{"_id": ID}, users)
									if err != nil {
										replyToMsg(messID, errStart+"send_shishs: update: from", peer, bot)
										errl.Println("e: ", err)
										return
									}
									err = docUpd(womb, wFil, users)
									if err != nil {
										replyToMsg(messID, errStart+"send_shishs: update: to", peer, bot)
										errl.Println("e: ", err)
										return
									}
									replyToMsg(messID,
										fmt.Sprintf("Вы успешно перевели %d шишей на счёт %s. Теперь у вас %d шишей при себе",
											amount, tWomb.Name, womb.Money), peer, bot,
									)
									sendMsg(fmt.Sprintf("Пользователь %s перевёл вам %d шишей. Теперь у вас %d шишей при себе",
										womb.Name, amount, tWomb.Money), ID, bot,
									)
								} else {
									replyToMsg(messID,
										fmt.Sprintf("Ошибка: пользователя с ID %d не найдено", ID), peer, bot)
								}
							} else {
								replyToMsg(messID,
									"Ошибка: количество переводимых шишей должно быть больше нуля", peer, bot)
							}
						} else {
							replyToMsg(messID,
								fmt.Sprintf("Ошибка: размер перевода (%d) должен быть меньше кол-ва ваших шишей при себе (%d)",
									amount, womb.Money), peer, bot,
							)
						}
					} else {
						if _, err := strconv.ParseInt(args[0], 10, 64); err == nil {
							replyToMsg(messID, "Ошибка: количество переводимых шишей должно быть больше нуля",
								peer, bot,
							)
						} else {
							replyToMsg(messID, "Ошибка: кол-во переводимых шишей быть числом", peer, bot)
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
					errl.Println("e: ", err)
					return
				}
				titles = []Title{}
				for cur.Next(ctx) {
					var nextOne Title
					err := cur.Decode(&nextOne)
					if err != nil {
						replyToMsg(messID, errStart+"update_data: titles_decode", peer, bot)
						errl.Println("e: ", err)
						return
					}
					titles = append(titles, nextOne)
				}
				cur.Close(ctx)
				imgsC = db.Collection("imgs")
				sendMsg("Успешно обновлено!", peer, bot)
				infl.Printf("DATA_UPDATE %d\n", peer)
				fmt.Printf("Data update by %d\n", peer)
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
					replyToMsg(messID, "Атака: аргументов должно быть больше одного", peer, bot)
					return
				}
				args := aargs[1:]
				al := len(args)
				switch args[0] {
				case "атака":
					replyToMsg(messID, strings.Repeat("атака ", 42), peer, bot)
				case "на":
					if al < 2 {
						sendMsg("Атака на: на кого?", peer, bot)
						return
					} else if al != 2 {
						replyToMsg(messID, "Атака на: слишком много аргументов", peer, bot)
						return
					} else if !isInUsers {
						replyToMsg(messID, "Вы не можете атаковать в виду остутствия вомбата", peer, bot)
						return
					} else if womb.Sleep {
						replyToMsg(messID, "Но вы же спите...", peer, bot)
						return
					}
					strID := args[1]
					var (
						ID    int64
						tWomb User
					)
					if is, isFrom := isInAttacks(from, attacks); isFrom {
						at, err := getAttackByWomb(from, true, attacks)
						if err != nil && err != errNoAttack {
							replyToMsg(messID, errStart+"attack: to: from_from: get_attack_by_womb", peer, bot)
							errl.Println("e: ", err)
							return
						}
						var aWomb User
						err = users.FindOne(ctx, bson.M{"_id": at.To}).Decode(&aWomb)
						if err != nil {
							replyToMsg(messID, errStart+"attack: to: find_attack_from", peer, bot)
							errl.Println("e: ", err)
							return
						}
						replyToMsgMD(messID, fmt.Sprintf(
							"Вы уже атакуете вомбата `%s`. Чтобы отозвать атаку, напишите `атака отмена`",
							aWomb.Name),
							peer, bot)
						return
					} else if is {
						at, err := getAttackByWomb(from, false, attacks)
						if err != nil && err != errNoAttack {
							replyToMsg(messID, errStart+"attack: to: from_to: get_attack_by_womb", peer, bot)
							errl.Println("e: ", err)
							return
						}
						var aWomb User
						err = users.FindOne(ctx, bson.M{"_id": at.From}).Decode(&aWomb)
						if err != nil {
							replyToMsg(messID, errStart+"attack: to: find_attack_to", peer, bot)
							errl.Println("e: ", err)
							return
						}
						replyToMsgMD(messID, fmt.Sprintf(
							"Вас уже атакует вомбат `%s`. Чтобы отклонить атаку, напишите `атака отмена`",
							aWomb.Name),
							peer, bot)
						return
					}
					if len([]rune(strID)) > 64 {
						replyToMsg(messID, "Слишком длинный никнейм", peer, bot)
						return
					} else if !isValidName(strID) {
						replyToMsg(messID, "нелегальный никнейм!!", peer, bot)
						return
					} else if rCount, err := users.CountDocuments(ctx,
						bson.M{"name": caseInsensitive(strID)}); err == nil && rCount != 0 {
						err = users.FindOne(ctx, bson.M{"name": caseInsensitive(strID)}).Decode(&tWomb)
						if err != nil {
							replyToMsg(messID, errStart+"attack: to: find_users_name", peer, bot)
							errl.Println("e: ", err)
							return
						}
						ID = tWomb.ID
					} else if err != nil {
						replyToMsg(messID, errStart+"attack: to: count_users_name", peer, bot)
						errl.Println("e: ", err)
						return
					} else {
						replyToMsg(messID, fmt.Sprintf("Пользователя с именем `%s` не найдено", strID),
							peer, bot)
						return
					}
					if ID == int64(from) {
						replyToMsg(messID, "„Основная борьба в нашей жизни — борьба с самим собой“ (c) какой-то философ", peer, bot)
						return
					}
					err = users.FindOne(ctx, bson.M{"_id": ID}).Decode(&tWomb)
					if err != nil {
						replyToMsg(messID, errStart+"attack: to: is_to_sleep", peer, bot)
						errl.Println("e: ", err)
						return
					}
					if tWomb.Sleep {
						replyToMsg(messID, fmt.Sprintf(
							"Вомбат %s спит. Его атаковать не получится",
							tWomb.Name), peer, bot)
						return
					} else if is, isFrom := isInAttacks(ID, attacks); isFrom {
						at, err := getAttackByWomb(ID, true, attacks)
						if err != nil && err != errNoAttack {
							replyToMsg(messID, errStart+"attack: to: to_from: get_attack_by_womb", peer, bot)
							errl.Println("e: ", err)
							return
						}
						var aWomb User
						err = users.FindOne(ctx, bson.M{"_id": at.To}).Decode(&aWomb)
						if err != nil {
							replyToMsg(messID, errStart+"attack: to: find_to_from", peer, bot)
							errl.Println("e: ", err)
							return
						}
						replyToMsgMD(messID, fmt.Sprintf(
							"%s уже атакует вомбата %s. Попросите %s решить данную проблему",
							strID, aWomb.Name, strID),
							peer, bot)
						return
					} else if is {
						at, err := getAttackByWomb(from, false, attacks)
						if err != nil && err != errNoAttack {
							replyToMsg(messID, errStart+"attack: to: to_to: get_attack_by_womb", peer, bot)
							errl.Println("e: ", err)
							return
						}
						var aWomb User
						err = users.FindOne(ctx, bson.M{"_id": at.From}).Decode(&aWomb)
						if err != nil {
							replyToMsg(messID, errStart+"attack: to: find_to_to", peer, bot)
							errl.Println("e: ", err)
							return
						}
						replyToMsg(messID, fmt.Sprintf(
							"Вомбат %s уже атакуется %s. Попросите %s решить данную проблему",
							strID, aWomb.Name, strID),
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
						errl.Println("e: ", err)
						return
					}
					replyToMsg(messID, fmt.Sprintf(
						"Вы отправили вомбата атаковать %s. Ждём ответа!\nОтменить можно командой `атака отмена`",
						tWomb.Name), peer, bot)
					sendMsg(fmt.Sprintf(
						"Ужас! Вас атакует %s. Предпримите какие-нибудь меры: отмените атаку (`атака отмена`) или примите (`атака принять`)",
						womb.Name), tWomb.ID, bot)
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
						if len([]rune(strID)) > 64 {
							replyToMsg(messID, "Слишком длинный никнейм", peer, bot)
							return
						} else if !isValidName(strID) {
							replyToMsg(messID, "Какой-то нелегальный никнейм", peer, bot)
							return
						} else if rCount, err := users.CountDocuments(ctx,
							bson.M{"name": caseInsensitive(strID)}); err == nil && rCount != 0 {
							var tWomb User
							err = users.FindOne(ctx, bson.M{"name": caseInsensitive(strID)}).Decode(&tWomb)
							if err != nil {
								replyToMsg(messID, errStart+"attack: find_users_name", peer, bot)
								errl.Println("e: ", err)
								return
							}
							ID = tWomb.ID
						} else if err != nil {
							replyToMsg(messID, errStart+"attack: status: count_users_name", peer, bot)
							errl.Println("e: ", err)
							return
						} else {
							replyToMsg(messID, fmt.Sprintf("Пользователя с никнеймом `%s` не найдено", strID), peer, bot)
							return
						}
					}
					var at Attack
					if is, isFrom := isInAttacks(ID, attacks); isFrom {
						a, err := getAttackByWomb(ID, true, attacks)
						if err != nil {
							replyToMsg(messID, errStart+"attack: status: to_at", peer, bot)
							errl.Println("e: ", err)
							return
						}
						at = a
					} else if is {
						a, err := getAttackByWomb(from, false, attacks)
						if err != nil {
							replyToMsg(messID, errStart+"attack: status: from_at", peer, bot)
							errl.Println("e: ", err)
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
						errl.Println("e: ", err)
						return
					}
					err = users.FindOne(ctx, bson.M{"_id": at.To}).Decode(&toWomb)
					if err != nil {
						replyToMsg(messID, errStart+"attack: status: find_twomb", peer, bot)
						errl.Println("e: ", err)
						return
					}
					replyToMsg(messID, fmt.Sprintf(
						"От: %s (%d)\nКому: %s (%d)\n",
						fromWomb.Name, fromWomb.ID,
						toWomb.Name, toWomb.ID,
					), peer, bot)
				case "отмена":
					if al > 1 {
						replyToMsg(messID, "атака отмена: слишком много аргументов", peer, bot)
					} else if !isInUsers {
						replyToMsg(messID, "какая атака, у тебя вобмата нет", peer, bot)
					}
					var at Attack
					if is, isFrom := isInAttacks(from, attacks); isFrom {
						a, err := getAttackByWomb(from, true, attacks)
						if err != nil {
							replyToMsg(messID, errStart+"attack: cancel: to_at", peer, bot)
							errl.Println("e: ", err)
							return
						}
						at = a
					} else if is {
						a, err := getAttackByWomb(from, false, attacks)
						if err != nil {
							replyToMsg(messID, errStart+"attack: cancel: from_at", peer, bot)
							errl.Println("e: ", err)
							return
						}
						at = a
					} else {
						replyToMsg(messID, "Атаки с вами не найдено...", peer, bot)
						return
					}
					_, err = attacks.DeleteOne(ctx, bson.M{"_id": at.ID})
					if err != nil {
						replyToMsg(messID, errStart+"attack: cancel: delete", peer, bot)
						errl.Println("e: ", err)
						return
					}
					can0, err := getImgs(imgsC, "cancel_0")
					if err != nil {
						replyToMsg(messID, errStart+"attack: cancel: get_imgs_0", peer, bot)
						errl.Println("e: ", err)
						return
					}
					can1, err := getImgs(imgsC, "cancel_1")
					if err != nil {
						replyToMsg(messID, errStart+"attack: cancel: get_imgs_1", peer, bot)
						errl.Println("e: ", err)
						return
					}
					if at.From == int64(from) {
						replyWithPhoto(messID, randImg(can0), "Вы отменили атаку", peer, bot)
						sendPhoto(randImg(can1),
							fmt.Sprintf("Вомбат %s решил вернуть вомбата домой. Вы свободны от атак",
								womb.Name), at.To, bot)
					} else {
						replyWithPhoto(messID, randImg(can0), "Вы отклонили атаку", peer, bot)
						sendPhoto(randImg(can1), fmt.Sprintf(
							"Вомбат %s вежливо отказал вам в войне. Вам пришлось забрать вомбата обратно. Вы свободны от атак",
							womb.Name), at.From, bot)
					}
				case "принять":
					if al > 2 {
						sendMsg("Атака принять: слишком много аргументов", peer, bot)
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
							replyToMsg(messID, errStart+"attack: accept: from_at", peer, bot)
							errl.Println("e: ", err)
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
						errl.Println("e: ", err)
						return
					} else if rCount < 1 {
						sendMsg("Ну ты чаво... Соперника не существует! Как вообще мы такое допустили?! (ответь на это командой /admin)",
							peer, bot)
						return
					}
					var tWomb User
					err = users.FindOne(ctx, bson.M{"_id": at.From}).Decode(&tWomb)
					if err != nil {
						replyToMsg(messID, errStart+"attack: accept: find_from", peer, bot)
						errl.Println("e: ", err)
						return
					}
					atimgs, err := getImgs(imgsC, "attacks")
					if err != nil {
						replyToMsg(messID, errStart+"attack: accept: imgs", peer, bot)
						errl.Println("e: ", err)
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
						errl.Println("e: ", err)
						return
					}
					err = docUpd(tWomb, bson.M{"_id": tWomb.ID}, users)
					if err != nil {
						replyToMsg(messID, errStart+"attack: accept: update_from", peer, bot)
						errl.Println("e: ", err)
						return
					}
					_, err = attacks.DeleteOne(ctx, bson.M{"_id": at.ID})
					if err != nil {
						replyToMsg(messID, errStart+"attack: accept: delete", peer, bot)
						errl.Println("e: ", err)
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
					errl.Println("e: ", err)
					return
				}
				sleep, err := getImgs(imgsC, "sleep")
				if err != nil {
					replyToMsg(messID, errStart+"go_sleep: get_imgs", peer, bot)
					errl.Println("e: ", err)
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
					errl.Println("e: ", err)
					return
				}
				unsleep, err := getImgs(imgsC, "unsleep")
				if err != nil {
					replyToMsg(messID, errStart+"unsleep: get_imgs", peer, bot)
					errl.Println("e: ", err)
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
					errl.Println("e: ", err)
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
					errl.Println("RROR err:rating: queue else")
					return
				}
				msg += ":\n"
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
					errl.Println("e: ", err)
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
						errl.Println("e: ", err)
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
						name := args[2]
						if len([]rune(name)) > 64 {
							replyToMsg(messID, "Слишком длинный никнейм", peer, bot)
							return
						} else if !isValidName(name) {
							replyToMsg(messID, "Нелегальное имя", peer, bot)
							return
						} else if rCount, err := users.CountDocuments(
							ctx, bson.M{"name": caseInsensitive(name)}); err == nil && rCount != 0 {
							err = users.FindOne(ctx, bson.M{"name": caseInsensitive(name)}).Decode(&tWomb)
							if err != nil {
								replyToMsg(messID, errStart+"bank: status: find_users_name", peer, bot)
								errl.Println("e: ", err)
								return
							}
							fil = bson.M{"_id": tWomb.ID}
							bCount, err := bank.CountDocuments(ctx, fil)
							if err != nil {
								replyToMsg(messID, errStart+"bank: status: count_banked", peer, bot)
								errl.Println("e: ", err)
								return
							}
							if bCount == 0 {
								replyToMsg(messID,
									fmt.Sprintf("Ошибка: вомбат с именем %s не зарегестрирован в вомбанке", name),
									peer, bot,
								)
								return
							}
						} else if err != nil {
							replyToMsg(messID, errStart+"bank: status: count_users_name", peer, bot)
							errl.Println("e: ", err)
							return
						} else {
							replyToMsg(messID, fmt.Sprintf("Ошибка: вомбата с именем %s не найдено", name), peer, bot)
							return
						}
					default:
						replyToMsg(messID, "Вомбанк статус: слишком много аргументов", peer, bot)
					}
					var b Banked
					err = bank.FindOne(ctx, fil).Decode(&b)
					if err != nil {
						replyToMsg(messID, errStart+"bank: status: find", peer, bot)
						errl.Println("e: ", err)
						return
					}
					replyToMsg(messID, fmt.Sprintf(
						"Вомбанк вомбата %s:\nНа счету: %d\nПри себе: %d",
						tWomb.Name, b.Money, tWomb.Money), peer, bot)
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
							errl.Println("e: ", err)
							return
						}
						womb.Money -= num
						b.Money += num
						err = docUpd(womb, wFil, users)
						if err != nil {
							replyToMsg(messID, errStart+"bank: put: upd_womb", peer, bot)
							errl.Println("e: ", err)
							return
						}
						err = docUpd(b, wFil, bank)
						if err != nil {
							replyToMsg(messID, errStart+"bank: put: upd_bank", peer, bot)
							errl.Println("e: ", err)
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
						errl.Println("e: ", err)
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
						errl.Println("e: ", err)
						return
					}
					err = docUpd(womb, wFil, users)
					if err != nil {
						replyToMsg(messID, errStart+"bank: put: upd_womb", peer, bot)
						errl.Println("e: ", err)
						return
					}
					replyToMsg(messID, fmt.Sprintf(
						"Вы сняли %d ш! Вомбосчёт: %d ш; При себе: %d ш",
						num, b.Money, womb.Money,
					), peer, bot)
				default:
					replyToMsg(messID, "Вомбанк: неизвестная команда: "+args[1], peer, bot)
				}
			} else if args := strings.Fields(txt); len(args) >= 1 && strings.ToLower(args[0]) == "клан" {
				if len(args) == 1 {
					replyToMsg(messID, "согласен", peer, bot)
					return
				} else if strings.ToLower(args[1]) == "клан" {
					replyToMsg(messID, strings.Repeat("клан ", 42), peer, bot)
					return
				} else if !isInUsers {
					if !(strings.ToLower(args[1]) == "статус" || (len(args) > 2 && strings.ToLower(args[1]) == "клан" &&
						strings.ToLower(args[2]) == "статус")) {
						replyToMsg(messID, "Кланы — приватная территория вомбатов. Как и всё в этом боте. У тебя же вомбата нет",
							peer, bot)
						return
					}
				}
				switch strings.ToLower(args[1]) {
				case "создать":
					if len(args) < 4 {
						replyToMsg(messID,
							"Клан создать: недостаточно аргументов. Синтаксис: клан создать "+
								"[тег (3-4 латинские буквы)] [имя (можно пробелы)]",
							peer, bot,
						)
						return
					} else if womb.Money < 25000 {
						replyToMsg(messID,
							"Ошибка: недостаточно шишей. Требуется 25'000 шишей при себе для создания клана "+
								fmt.Sprintf("(У вас их при себе %d)", womb.Money),
							peer, bot,
						)
						return
					} else if l := len([]rune(args[2])); !(l >= 3 && l <= 5) {
						replyToMsg(messID, "Слишком длинный тэг!", peer, bot)
						return
					} else if !isValidTag(args[2]) {
						replyToMsg(messID, "Нелегальный тэг(", peer, bot)
						return
					} else if name := strings.Join(args[3:], " "); len([]rune(name)) > 64 {
						replyToMsg(messID, "Слишком длинное имя! Оно должно быть максимум 64 символов",
							peer, bot,
						)
						return
					} else if len([]rune(name)) < 2 {
						replyToMsg(messID, "Слишком короткое имя! Оно должно быть минимум 3 символа",
							peer, bot,
						)
						return
					}
					tag, name := strings.ToLower(args[2]), strings.Join(args[3:], " ")
					if rCount, err := clans.CountDocuments(ctx,
						bson.M{"_id": caseInsensitive(tag)}); err != nil {
						replyToMsg(messID, errStart+"clan: new: count_tag", peer, bot)
						errl.Println("e: ", err)
						return
					} else if rCount != 0 {
						replyToMsg(messID, fmt.Sprintf(
							"Ошибка: клан с тегом `%s` уже существует",
							tag),
							peer, bot,
						)
						return
					}
					if rCount, err := clans.CountDocuments(ctx,
						bson.M{"members": from}); err != nil {
						replyToMsg(messID, errStart+"clan: new: count_members", peer, bot)
						errl.Println("e: ", err)
						return
					} else if rCount != 0 {
						replyToMsg(messID,
							"Ошибка: вы уже состоите в клане. Напишите `клан выйти`, чтобы выйти из него",
							peer, bot,
						)
						return
					}
					womb.Money -= 25000
					err = docUpd(womb, wFil, users)
					if err != nil {
						replyToMsg(messID, errStart+"clan: new: update_money", peer, bot)
						errl.Println("e: ", err)
						return
					}
					nclan := Clan{
						Tag:     strings.ToUpper(tag),
						Name:    name,
						Money:   100,
						Leader:  from,
						Members: []int64{from},
					}
					_, err := clans.InsertOne(ctx, &nclan)
					if err != nil {
						replyToMsg(messID, errStart+"clan: new: insert", peer, bot)
						errl.Println("e: ", err)
						return
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
						err = docUpd(womb, wFil, users)
						if err != nil {
							replyToMsg(messID, errStart+"clan: transfer: delete_title", peer, bot)
							errl.Println("e: ", err)
							return
						}
					}
					replyToMsg(messID,
						fmt.Sprintf("Клан `%s` успешно создан! У вас взяли 25'000 шишей", name),
						peer, bot,
					)
				case "вступить":
					if len(args) != 3 {
						replyToMsg(messID, "Клан вступить: слишком мало или много аргументов! "+
							"Синтаксис: клан вступить [тэг клана]",
							peer, bot,
						)
						return
					} else if womb.Money < 1000 {
						replyToMsg(messID, "Клан вступить: недостаточно шишей (надо минимум 1000 ш)",
							peer, bot,
						)
						return
					} else if rCount, err := clans.CountDocuments(ctx,
						bson.M{"members": from}); err != nil {
						replyToMsg(messID, errStart+"clan: join: count_members", peer, bot)
						errl.Println("e: ", err)
						return
					} else if rCount != 0 {
						replyToMsg(messID,
							"Ошибка: вы уже состоите в клане. Напишите `клан выйти`, чтобы выйти из него",
							peer, bot,
						)
						return
					} else if l := len([]rune(args[2])); !(l >= 3 && l <= 5) {
						replyToMsg(messID, "Слишком длинный или короткий тег :)", peer, bot)
						return
					} else if !isValidTag(args[2]) {
						replyToMsg(messID, "Тег нелгальный(", peer, bot)
						return
					} else if rCount, err := clans.CountDocuments(ctx,
						bson.M{"_id": strings.ToUpper(args[2])}); err != nil {
						replyToMsg(messID, errStart+"clan: join: count_tag", peer, bot)
						errl.Println("e: ", err)
						return
					} else if rCount == 0 {
						replyToMsg(messID,
							fmt.Sprintf("Ошибка: клана с тегом `%s` не существует", args[2]),
							peer, bot,
						)
						return
					}
					var jClan Clan
					err = clans.FindOne(ctx, bson.M{"_id": strings.ToUpper(args[2])}).Decode(&jClan)
					if err != nil {
						replyToMsg(messID, errStart+"clan: join: find_clan", peer, bot)
						errl.Println("e: ", err)
						return
					}
					if len(jClan.Members) >= 7 {
						replyToMsg(messID, "Ошибка: в клане слишком много игроков :(", peer, bot)
						return
					}
					womb.Money -= 1000
					err = docUpd(womb, wFil, users)
					if err != nil {
						replyToMsg(messID, errStart+"clan: join: update_user", peer, bot)
						errl.Println("e: ", err)
						return
					}
					jClan.Members = append(jClan.Members, from)
					err = docUpd(jClan, bson.M{"_id": strings.ToUpper(args[2])}, clans)
					if err != nil {
						replyToMsg(messID, errStart+"clan: join: update_clan", peer, bot)
						errl.Println("e: ", err)
						return
					}
					replyToMsg(messID, "Отлично, вы присоединились! У вас взяли 1000 шишей",
						peer, bot,
					)
					sendMsg(
						fmt.Sprintf("В ваш клан вступил вомбат `%s`", womb.Name),
						jClan.Leader, bot,
					)
				case "назначить":
					if len(args) == 2 {
						replyToMsg(messID, "конечно", peer, bot)
						return
					}
					switch args[2] {
					case "назначить":
						replyToMsg(messID, strings.Repeat("назначить", 42), peer, bot)
						return
					case "лидера":
						fallthrough
					case "лидером":
						fallthrough
					case "лидер":
						replyToMsg(messID, "Используйте \"клан передать [имя]\" вместо данной команды", peer, bot)
					case "казначея":
						fallthrough
					case "казначеем":
						fallthrough
					case "казначей":
						if len(args) != 4 {
							replyToMsg(messID, "Слишком много или мало аргументов", peer, bot)
							return
						} else if !isInUsers {
							replyToMsg(messID, "Кланы — приватная территория вомбатов. У тебя вомбата нет.", peer, bot)
							return
						}
						if c, err := clans.CountDocuments(ctx, bson.M{"leader": from}); err != nil {
							replyToMsg(messID, errStart+"count_leader_clan", peer, bot)
							errl.Println("e: ", err)
							return
						} else if c == 0 {
							replyToMsg(messID, "Вы не состоите ни в одном клане либо не являетесь лидером клана", peer, bot)
							return
						}
						var sClan Clan
						if err := clans.FindOne(ctx, bson.M{"leader": from}).Decode(&sClan); err != nil {
							replyToMsg(messID, errStart+"find_leader_clan", peer, bot)
							errl.Println("e: ", err)
							return
						}
						lbid := sClan.Banker
						name := args[3]
						if c, err := users.CountDocuments(ctx, bson.M{"name": caseInsensitive(name)}); err != nil {
							replyToMsg(messID, errStart+"count_new_banker", peer, bot)
							errl.Println("e: ", err)
							return
						} else if c == 0 {
							replyToMsg(messID, "Вомбата с таким ником не найдено", peer, bot)
							return
						}
						var (
							nb User
						)
						if err := users.FindOne(ctx, bson.M{"name": caseInsensitive(name)}).Decode(&nb); err != nil {
							replyToMsg(messID, errStart+"find_new_banker", peer, bot)
							errl.Println("e: ", err)
							return
						}
						var is bool
						for _, id := range sClan.Members {
							if id == nb.ID {
								is = true
								break
							}
						}
						if !is {
							replyToMsg(messID, "Данный вобат не состоит в Вашем клане", peer, bot)
							return
						}
						sClan.Banker = nb.ID
						if err := docUpd(sClan, bson.M{"_id": sClan.Tag}, clans); err != nil {
							replyToMsg(messID, errStart+"update_clan", peer, bot)
							errl.Println("e: ", err)
							return
						}
						replyToMsg(messID, "Казначей успешно изменён! Теперь это "+nb.Name, peer, bot)
						if nb.ID != from {
							sendMsg("Вы стали казначеем в клане `"+sClan.Name+"` ["+sClan.Tag+"]", nb.ID, bot)
						}
						if lbid != from && lbid != 0 {
							sendMsg("Вы больше не казначеем... (в клане `"+sClan.Name+"` ["+sClan.Tag+"])", lbid, bot)
						}
					default:
						replyToMsg(messID, "Не знаю такой роли в клане(", peer, bot)
						return
					}
				case "передать":
					if len(args) != 3 {
						replyToMsg(messID, "Ошибка: слишком много или мало аргументов. "+
							"Синтаксис: клан передать [ник]",
							peer, bot,
						)
						return
					} else if rCount, err := clans.CountDocuments(ctx,
						bson.M{"members": from}); err != nil {
						replyToMsg(messID, errStart+"clan: transfer: count_members_from", peer, bot)
						errl.Println("e: ", err)
						return
					} else if rCount == 0 {
						replyToMsg(messID, "Ошибка: вы не состоите ни в каком клане", peer, bot)
						return
					} else if rCount, err := clans.CountDocuments(ctx,
						bson.M{"leader": from}); err != nil {
						replyToMsg(messID, errStart+"clan: transfer: count_leader_from", peer, bot)
						errl.Println("e: ", err)
						return
					} else if rCount == 0 {
						replyToMsg(messID, "Ошибка: вы не лидер!!!11!!!", peer, bot)
						return
					} else if len([]rune(args[2])) > 64 {
						replyToMsg(messID, "Ошибка: слишком длинный ник", peer, bot)
						return
					} else if rCount, err := users.CountDocuments(ctx,
						bson.M{"name": caseInsensitive(args[2])}); err != nil {
						replyToMsg(messID, errStart+"", peer, bot)
						errl.Println("e: ", err)
						return
					} else if rCount == 0 {
						replyToMsg(messID,
							fmt.Sprintf("Ошибка: пользователя с ником `%s` не существует", args[2]),
							peer, bot,
						)
						return
					}
					var newLead User
					err = users.FindOne(ctx, bson.M{"name": caseInsensitive(args[2])}).Decode(&newLead)
					if err != nil {
						replyToMsg(messID, errStart+"clan: transfer: find_new_lead", peer, bot)
						errl.Println("e: ", err)
						return
					}
					if strings.ToLower(args[2]) == strings.ToLower(womb.Name) {
						replyToMsg(messID, "Но ты и так лидер...", peer, bot)
						return
					} else if rCount, err := clans.CountDocuments(ctx,
						bson.M{"members": newLead.ID}); err != nil {
						replyToMsg(messID, errStart+"clan: transfer: count_new_lead_clan", peer, bot)
						errl.Println("e: ", err)
						return
					} else if rCount == 0 {
						replyToMsg(messID,
							fmt.Sprintf("Ошибка: вомбат `%s` не состоит ни в одном клане", newLead.Name),
							peer, bot,
						)
						return
					}
					var uClan Clan
					err = clans.FindOne(ctx, bson.M{"leader": from}).Decode(&uClan)
					if err != nil {
						replyToMsg(messID, errStart+"clan: transfer: find_leaders_clan", peer, bot)
						errl.Println("e: ", err)
						return
					}
					var isIn bool = false
					for _, id := range uClan.Members {
						if id == newLead.ID {
							isIn = true
							break
						}
					}
					if !isIn {
						replyToMsg(messID,
							fmt.Sprintf("Ошибка: вы и %s состоите в разных кланах", newLead.Name),
							peer, bot,
						)
						return
					}
					uClan.Leader = newLead.ID
					err = docUpd(uClan, bson.M{"_id": uClan.Tag}, clans)
					if err != nil {
						replyToMsg(messID, errStart+"clan: transfer: update", peer, bot)
						errl.Println("e: ", err)
						return
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
							replyToMsg(messID, errStart+"clan: transfer: delete_title", peer, bot)
							errl.Println("e: ", err)
							return
						}
					}
					if !hasTitle(5, newLead.Titles) {
						newLead.Titles = append(newLead.Titles, 5)
						err = docUpd(newLead, bson.M{"_id": newLead.ID}, users)
						if err != nil {
							replyToMsg(messID, errStart+"clan: transfer: add_title", peer, bot)
							errl.Println("e: ", err)
							return
						}
					}
					replyToMsg(messID,
						fmt.Sprintf("Отлично! Вомбат `%s` теперь главный в клане `%s`",
							newLead.Name, uClan.Tag),
						peer, bot,
					)
					sendMsg("Вам передали права на клан!", newLead.ID, bot)
				case "выйти":
					if len(args) != 2 {
						replyToMsg(messID,
							"Ошибка: слишком много или мало аргументов. Синтаксис: клан выйти",
							peer, bot,
						)
						return
					} else if rCount, err := clans.CountDocuments(ctx,
						bson.M{"members": from}); err != nil {
						replyToMsg(messID, errStart+"clan: quit: count_clan", peer, bot)
						errl.Println("e: ", err)
						return
					} else if rCount == 0 {
						replyToMsg(messID, "Клан выйти: вы не состоите ни в одном клане", peer, bot)
						return
					}
					var uClan Clan
					err = clans.FindOne(ctx, bson.M{"members": from}).Decode(&uClan)
					if err != nil {
						replyToMsg(messID, errStart+"clan: quit: find_clan", peer, bot)
						errl.Println("e: ", err)
						return
					}
					if len(uClan.Members) == 1 {
						_, err = clans.DeleteOne(ctx, bson.M{"_id": uClan.Tag})
						if err != nil {
							replyToMsg(messID, errStart+"clan: quit: delete", peer, bot)
							errl.Println("e: ", err)
							return
						}
						if uClan.Leader == from {
							if hasTitle(5, womb.Titles) {
								newTitles := []uint16{}
								for _, id := range womb.Titles {
									if id == 5 {
										continue
									}
									newTitles = append(newTitles, id)
								}
								womb.Titles = newTitles
								err = docUpd(womb, wFil, users)
								if err != nil {
									replyToMsg(messID, errStart+"clan: quit: delete_title", peer, bot)
									errl.Println("e: ", err)
									return
								}
							}
						}
						replyToMsg(messID, "Так как вы были одни в клане, то клан удалён", peer, bot)
						return
					} else if uClan.Leader == from {
						replyToMsg(messID, "Клан выйти: вы лидер. Передайте кому-либо ваши права", peer, bot)
						return
					}
					newMembers := []int64{}
					for _, id := range uClan.Members {
						if id == from {
							continue
						}
						newMembers = append(newMembers, id)
					}
					var (
						rep    string = "Вы вышли из клана. Вы свободны!"
						msgtol string = "Вомбат `" + womb.Name + "` вышел из клана."
					)
					uClan.Members = newMembers
					if uClan.Banker == from && uClan.Leader != uClan.Banker {
						uClan.Banker = uClan.Leader
						rep += "\nБанкиром вместо вас стал лидер клана."
						msgtol += "\nТак как этот вомбат был банкиром, Вы стали банкиром клана."
					}
					err = docUpd(uClan, bson.M{"_id": uClan.Tag}, clans)
					if err != nil {
						replyToMsg(messID, errStart+"clan: quit: update", peer, bot)
						errl.Println("e: ", err)
						return
					}
					replyToMsg(messID, rep, peer, bot)
					sendMsg(msgtol, uClan.Leader, bot)
				case "статус":
					if len(args) > 3 {
						replyToMsg(messID,
							"Клан статус: слишком много аргументов! Синтаксис: клан статус ([тег])",
							peer, bot,
						)
						return
					}
					var sClan Clan
					if len(args) == 2 {
						if !isInUsers {
							replyToMsg(messID,
								"Вы не имеете вомбата. Соответственно, вы не состоите в ни в одном вомбоклане",
								peer, bot,
							)
							return
						} else if rCount, err := clans.CountDocuments(ctx,
							bson.M{"members": from}); err != nil {
							replyToMsg(messID, errStart+"clan: status: count_from_clan", peer, bot)
							errl.Println("e: ", err)
							return
						} else if rCount == 0 {
							replyToMsg(messID, "Клан статус: вы не состоите ни в одном клане", peer, bot)
							return
						}
						err = clans.FindOne(ctx, bson.M{"members": from}).Decode(&sClan)
						if err != nil {
							replyToMsg(messID, errStart+"clan: status: find_from_clan", peer, bot)
							errl.Println("e: ", err)
							return
						}
					} else {
						if l := len([]rune(args[2])); !(l >= 3 && l <= 5) {
							replyToMsg(messID, "Ошибка: слишком длинный или короткий тег", peer, bot)
							return
						} else if !isValidTag(args[2]) {
							replyToMsg(messID, "Ошибка: тег нелегален", peer, bot)
							return
						} else if rCount, err := clans.CountDocuments(ctx,
							bson.M{"_id": caseInsensitive(args[2])}); err != nil {
							replyToMsg(messID, errStart+"clan: status: count_clan", peer, bot)
							errl.Println("e: ", err)
							return
						} else if rCount == 0 {
							replyToMsg(messID,
								fmt.Sprintf("Ошибка: клана с тегом `%s` не существует", args[2]),
								peer, bot,
							)
							return
						}
						err = clans.FindOne(ctx, bson.M{"_id": caseInsensitive(args[2])}).Decode(&sClan)
						if err != nil {
							replyToMsg(messID, errStart+"clan: status: find_clan", peer, bot)
							errl.Println("e: ", err)
							return
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
					for i, id := range sClan.Members {
						if id == sClan.Leader && i != 0 {
							continue
						}
						if rCount, err := users.CountDocuments(ctx,
							bson.M{"_id": id}); err != nil {
							replyToMsg(messID, errStart+"clan: status: count_user", peer, bot)
							errl.Println("e: ", err)
							return
						} else if rCount == 0 {
							msg += " - Вомбат не найден :("
							lost++
							continue
						} else {
							err = users.FindOne(ctx, bson.M{"_id": id}).Decode(&tWomb)
							if err != nil {
								replyToMsg(messID, errStart+"clan: status: find_user", peer, bot)
								errl.Println("e: ", err)
								return
							}
							msg += fmt.Sprintf("        %d. %s", i+1, tWomb.Name)
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
					replyToMsg(messID, msg, peer, bot)
				case "атака":
					if len(args) == 2 {
						replyToMsg(messID, "атака", peer, bot)
					}
					switch strings.ToLower(args[2]) {
					case "атака":
						replyToMsg(messID, strings.Repeat("атака ", 42), peer, bot)
					case "на":
						if len(args) == 3 {
							sendMsg("Атака на: на кого?", peer, bot)
							return
						} else if len(args) > 4 {
							replyToMsg(messID, "Атака на: слишком много аргументов", peer, bot)
							return
						} else if rCount, err := clans.CountDocuments(ctx,
							bson.M{"members": from}); err != nil {
							replyToMsg(messID, errStart+"clan: attack: to: count_from_clan", peer, bot)
							errl.Println("e: ", err)
							return
						} else if rCount == 0 {
							replyToMsg(messID, "Вы не состоите ни в одном клане", peer, bot)
							return
						} else if rCount, err := clans.CountDocuments(ctx,
							bson.M{"leader": from}); err != nil {
							replyToMsg(messID, errStart+"clan: attack: to: count_leader_clan", peer, bot)
							errl.Println("e: ", err)
							return
						} else if rCount == 0 {
							replyToMsg(messID, "Вы не являетесь лидером клана, в котором состоите", peer, bot)
							return
						}
						var fromClan Clan
						err = clans.FindOne(ctx, bson.M{"leader": from}).Decode(&fromClan)
						if err != nil {
							replyToMsg(messID, errStart+"clan: attack: to: find_from_clan", peer, bot)
							errl.Println("e: ", err)
							return
						} else if ok, from := isInClattacks(fromClan.Tag, clattacks); from {
							replyToMsg(messID, "Вы уже нападаете на другой клан", peer, bot)
							return
						} else if ok {
							replyToMsg(messID, "На вас уже нападают)", peer, bot)
							return
						}
						tag := strings.ToUpper(args[3])
						if len([]rune(tag)) > 64 {
							replyToMsg(messID, "Ошибка: слишком длинный тег!", peer, bot)
							return
						} else if !isValidTag(tag) {
							replyToMsg(messID, "Нелегальный тег", peer, bot)
							return
						} else if fromClan.Tag == tag {
							replyToMsg(messID, "гений", peer, bot)
							return
						} else if rCount, err := clans.CountDocuments(ctx,
							bson.M{"_id": tag}); err != nil {
							replyToMsg(messID, errStart+"clan: attack: to: count_clans", peer, bot)
							errl.Println("e: ", err)
							return
						} else if rCount == 0 {
							replyToMsg(messID, "Ошибка: клана с таким тегом не найдено", peer, bot)
						} else if ok, from := isInClattacks(tag, clattacks); from {
							replyToMsg(messID, "Клан "+tag+" уже атакует кого-то", peer, bot)
							return
						} else if ok {
							replyToMsg(messID, "Клан "+tag+" уже атакуется", peer, bot)
							return
						}
						var toClan Clan
						err = clans.FindOne(ctx, bson.M{"_id": tag}).Decode(&toClan)
						if err != nil {
							replyToMsg(messID, errStart+"clan: attack: to: find_to", peer, bot)
							errl.Println("e: ", err)
							return
						}
						newClat := Clattack{
							ID:   fromClan.Tag + "_" + tag,
							From: fromClan.Tag,
							To:   tag,
						}
						_, err := clattacks.InsertOne(ctx, newClat)
						if err != nil {
							replyToMsg(messID, errStart+"clan: attack: to: insert", peer, bot)
							errl.Println("e: ", err)
							return
						}
						replyToMsg(messID, "Отлично! Вы отправили вомбатов ждать согласия на вомбой",
							peer, bot)
						sendMsg("АААА!!! НА ВАС НАПАЛ КЛАН "+fromClan.Tag+". предпримите что-нибудь(",
							toClan.Leader, bot)
					case "отмена":
						if len(args) != 3 {
							replyToMsg(messID, "Клан атака отмена: слишком много аргументов", peer, bot)
							return
						} else if rCount, err := clans.CountDocuments(ctx,
							bson.M{"members": from}); err != nil {
							replyToMsg(messID, errStart+"clan: attack: cancel: count_from_clan", peer, bot)
							errl.Println("e: ", err)
							return
						} else if rCount == 0 {
							replyToMsg(messID, "Ошибка: вы не состоите ни в одном клане", peer, bot)
							return
						} else if rCount, err := clans.CountDocuments(ctx,
							bson.M{"leader": from}); err != nil {
							replyToMsg(messID, errStart+"clan: attack: cancel: count_leader_clan", peer, bot)
							errl.Println("e: ", err)
							return
						} else if rCount == 0 {
							replyToMsg(messID, "Ошибка: вы не являетесь лидером в своём клане", peer, bot)
							return
						}
						var cClan Clan
						err = clans.FindOne(ctx, bson.M{"leader": from}).Decode(&cClan)
						if err != nil {
							replyToMsg(messID, errStart+"clan: attack: cancel: find_from_clan", peer, bot)
							errl.Println("e: ", err)
							return
						}
						is, isfr := isInClattacks(cClan.Tag, clattacks)
						if !is {
							replyToMsg(messID, "Вы никого не атакуете и никем не атакуетесь. Вам нечего отменять :)",
								peer, bot)
							return
						}
						var clat Clattack
						err = clattacks.FindOne(ctx, bson.M{func(isfr bool) string {
							if isfr {
								return "from"
							}
							return "to"
						}(isfr): cClan.Tag}).Decode(&clat)
						if err != nil {
							replyToMsg(messID, errStart+"clan: attack: cancel: find_clattack", peer, bot)
							errl.Println("e: ", err)
							return
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
							replyToMsg(messID, errStart+"clan: attack: cancel: count_other_clan", peer, bot)
							errl.Println("e: ", err)
							return
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
								replyToMsg(messID, errStart+"clan: attack: cancel: find_other_clan", peer, bot)
								errl.Println("e: ", err)
								return
							}
						}
						_, err = clattacks.DeleteOne(ctx, bson.M{"to": clat.To})
						if err != nil {
							replyToMsg(messID, errStart+"clan: attack: cancel: delete", peer, bot)
							errl.Println("e: ", err)
							return
						}
						can0, err := getImgs(imgsC, "cancel_0")
						if err != nil {
							replyToMsg(messID, errStart+"clan: attack: cancel: get_imgs_0", peer, bot)
							errl.Println("e: ", err)
							return
						}
						var can1 Imgs
						if send {
							can1, err = getImgs(imgsC, "cancel_1")
							if err != nil {
								replyToMsg(messID, errStart+"clan: attack: cancel: get_imgs_1", peer, bot)
								errl.Println("e: ", err)
								return
							}
						}
						replyWithPhoto(messID, randImg(can0), "Вы "+func(isfr bool) string {
							if isfr {
								return "отменили"
							}
							return "отклонили"
						}(isfr)+" клановую атаку", peer, bot)
						if send {
							sendPhoto(randImg(can1), "Вашу клановую атаку "+func(isfr bool) string {
								if isfr {
									return "отменили"
								}
								return "отклонили"
							}(isfr)+")", oClan.Leader, bot)
						}
					case "принять":
						if len(args) != 3 {
							replyToMsg(messID, "Слишком много аргументов", peer, bot)
							return
						} else if rCount, err := clans.CountDocuments(ctx,
							bson.M{"members": from}); err != nil {
							replyToMsg(messID, errStart+"clan: attack: accept: count_to_clan", peer, bot)
							errl.Println("e: ", err)
							return
						} else if rCount == 0 {
							replyToMsg(messID, "Вы не состоите ни в одном клане", peer, bot)
							return
						} else if rCount, err := clans.CountDocuments(ctx,
							bson.M{"leader": from}); err != nil {
							replyToMsg(messID, errStart+"clan: attack: accept: count_leader_clan", peer, bot)
							errl.Println("e: ", err)
							return
						} else if rCount == 0 {
							replyToMsg(messID, "Вы не являетесь лидером клана", peer, bot)
							return
						}
						var toClan Clan
						err = clans.FindOne(ctx, bson.M{"leader": from}).Decode(&toClan)
						if err != nil {
							replyToMsg(messID, errStart+"clan: attack: accept: find_to_clan", peer, bot)
							errl.Println("e: ", err)
							return
						}
						if is, isfr := isInClattacks(toClan.Tag, clattacks); !is {
							replyToMsg(messID, "Ваш клан не атакуется/не атакует", peer, bot)
							return
						} else if isfr {
							replyToMsg(messID, "Принимать вомбой может только атакуемая сторона", peer, bot)
						}
						var clat Clattack
						err = clattacks.FindOne(ctx, bson.M{"to": toClan.Tag}).Decode(&clat)
						if err != nil {
							replyToMsg(messID, errStart+"clan: attacl: start: find_clattack", peer, bot)
							errl.Println("e: ", err)
							return
						}
						if rCount, err := clans.CountDocuments(ctx, bson.M{"_id": clat.From}); err != nil {
							replyToMsg(messID, errStart+"clan: attack: accept: count_from_clan", peer, bot)
							errl.Println("e: ", err)
							return
						} else if rCount == 0 {
							replyToMsg(messID, "Ошибка: атакующего клана не существует!", peer, bot)
							return
						}
						var frClan Clan
						err = clans.FindOne(ctx, bson.M{"_id": clat.From}).Decode(&frClan)
						if err != nil {
							replyToMsg(messID, errStart+"clan: attack: accept: find_from_clan", peer, bot)
							errl.Println("e: ", err)
							return
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
									replyToMsg(messID, errStart+"clan: attack: accept: count_user", peer, bot)
									errl.Println("e: ", err)
									return
								} else if rCount == 0 {
									lost++
									continue
								} else {
									err = users.FindOne(ctx, bson.M{"_id": id}).Decode(&tWomb)
									if err != nil {
										replyToMsg(messID, errStart+"clan: attack: accept: find_user", peer, bot)
										errl.Println("e: ", err)
										return
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
									replyToMsg(messID,
										"Ошибка: у клана ["+sClan.Tag+"] все вомбаты потеряны( ответьте командой /admin",
										peer, bot,
									)
									return
								}
							}
						}
						atimgs, err := getImgs(imgsC, "attacks")
						if err != nil {
							replyToMsg(messID, errStart+"clan: attack: accept: imgs", peer, bot)
							errl.Println("e: ", err)
							return
						}
						im := randImg(atimgs)
						ph1 := replyWithPhoto(messID, im, "", peer, bot)
						ph2 := sendPhoto(im, "", frClan.Leader, bot)
						war1 := replyToMsg(ph1, "Да начнётся вомбой!", peer, bot)
						war2 := replyToMsg(ph2, fmt.Sprintf(
							"АААА ВАЙНААААА!!!\n Вомбат %s всё же принял ваше предложение",
							womb.Name), frClan.Leader, bot,
						)
						time.Sleep(5 * time.Second)
						h1, h2 := int(toclwar.Health), int(frclwar.Health)
						for _, round := range []int{1, 2, 3} {
							f1 := uint32(2 + rand.Intn(int(toclwar.Force-1)))
							f2 := uint32(2 + rand.Intn(int(frclwar.Force-1)))
							editMsg(war1, fmt.Sprintf(
								"РАУНД %d\n\n[%s]:\n - здоровье: %d\n - Ваш удар: %d\n\n[%s]:\n - здоровье: %d",
								round, toClan.Tag, h1, f1, frClan.Tag, h2), peer, bot)
							editMsg(war2, fmt.Sprintf(
								"РАУНД %d\n\n[%s]:\n - здоровье: %d\n - Ваш удар: %d\n\n[%s]:\n - здоровье: %d",
								round, frClan.Tag, h2, f2, toClan.Tag, h1), frClan.Leader, bot)
							time.Sleep(3 * time.Second)
							h1 -= int(f2)
							h2 -= int(f1)
							editMsg(war1, fmt.Sprintf(
								"РАУНД %d\n\n[%s]\n - здоровье: %d\n - Ваш удар: %d\n\n[%s]:\n - здоровье: %d\n - 💔 удар: %d",
								round, toClan.Tag, h1, f1, frClan.Tag, h2, f2), peer, bot)
							editMsg(war2, fmt.Sprintf(
								"РАУНД %d\n\n[%s]:\n - здоровье: %d\n - Ваш удар: %d\n\n[%s]:\n - здоровье: %d\n - 💔 удар: %d",
								round, frClan.Tag, h2, f2, toClan.Tag, h1, f1), frClan.Leader, bot)
							time.Sleep(5 * time.Second)
							if int(h2)-int(f1) <= 5 && int(h1)-int(f2) <= 5 {
								editMsg(war1,
									"Оба клана сдохли!!!)\nВаши характеристики не поменялись, но зато да.",
									peer, bot)
								editMsg(war2,
									"Оба клана сдохли!!!)\nВаши характеристики не поменялись, но зато да.",
									frClan.Leader, bot)
								time.Sleep(5 * time.Second)
								break
							} else if int(h2)-int(f1) <= 5 {
								editMsg(war1, fmt.Sprintf(
									"В раунде %d благодаря силе участников победил клан...",
									round), peer, bot)
								editMsg(war2, fmt.Sprintf(
									"В раунде %d благодаря лишению у другого здоровья победил клан...",
									round), frClan.Leader, bot)
								time.Sleep(3 * time.Second)
								toClan.XP += 10
								editMsg(war1, fmt.Sprintf(
									"Победил клан `%s` [%s]!!!\nВы получили 10 XP, теперь их у вас %d",
									toClan.Name, toClan.Tag, toClan.XP), peer, bot)
								editMsg(war2, fmt.Sprintf(
									"Победил клан `%s` [%s]!!!\nВаше состояние не изменилось)",
									toClan.Name, toClan.Tag), frClan.Leader, bot)
								break
							} else if int(h1)-int(f2) <= 5 {
								editMsg(war1, fmt.Sprintf(
									"В раунде %d благодаря силе участников победил клан...",
									round), peer, bot)
								editMsg(war2, fmt.Sprintf(
									"В раунде %d благодаря лишению у другого здоровья победил клан...",
									round), frClan.Leader, bot)
								time.Sleep(3 * time.Second)
								frClan.XP += 10
								editMsg(war2, fmt.Sprintf(
									"Победил клан `%s` %s!!!\nВы получили 10 XP, теперь их у Вас %d",
									frClan.Name, frClan.Tag, frClan.XP), frClan.Leader, bot)
								womb.Health = 5
								womb.Money = 50
								editMsg(war1, fmt.Sprintf(
									"Победил клан `%s` [%s]!!!\nВаше состояние не изменилось)",
									frClan.Name, frClan.Tag), peer, bot)
								break
							} else if round == 3 {
								frClan.XP += 10
								if h1 < h2 {
									editMsg(war2, fmt.Sprintf(
										"И победил клан `%s` %s!!!\nВы получили 10 XP, теперь их у Вас %d",
										frClan.Name, frClan.Tag, frClan.XP), frClan.Leader, bot)
									editMsg(war1, fmt.Sprintf(
										"И победил клан `%s` [%s]!!!\nВаше состояние не изменилось)",
										frClan.Name, frClan.Tag), peer, bot)
								} else {
									toClan.XP += 10
									editMsg(war1, fmt.Sprintf(
										"Победил клан `%s` [%s]!!!\nВы получили 10 XP, теперь их у вас %d",
										toClan.Name, toClan.Tag, toClan.XP), peer, bot)
									editMsg(war2, fmt.Sprintf(
										"Победил клан `%s` [%s]!!!\nВаше состояние не изменилось)",
										toClan.Name, toClan.Tag), frClan.Leader, bot)
								}
							}
							err = docUpd(toClan, bson.M{"_id": toClan.Tag}, clans)
							if err != nil {
								replyToMsg(messID, errStart+"clan: attack: accept: update_to", peer, bot)
								errl.Println("e: ", err)
								return
							}
							err = docUpd(frClan, bson.M{"_id": frClan.Tag}, clans)
							if err != nil {
								replyToMsg(messID, errStart+"clan: attack: accept: update_from", peer, bot)
								errl.Println("e: ", err)
								return
							}
							_, err = clattacks.DeleteOne(ctx, bson.M{"_id": clat.ID})
							if err != nil {
								replyToMsg(messID, errStart+"clan: attack: accept: delete", peer, bot)
								errl.Println("e: ", err)
								return
							}
						}
					case "статус":
						var sClan Clan
						switch len(args) - 3 {
						case 0:
							if !isInUsers {
								replyToMsg(messID, "Вы не имеете вомбата => Вы не состоите ни в одном клане. Добавьте тег.", peer, bot)
								return
							}
							if c, err := clans.CountDocuments(ctx, bson.M{"members": from}); err != nil {
								replyToMsg(messID, errStart+"count_from_clan", peer, bot)
								errl.Println("e: ", err)
								return
							} else if c == 0 {
								replyToMsg(messID, "Вы не состоите ни в одном клане. Добавьте тег.", peer, bot)
								return
							}
							if err := clans.FindOne(ctx, bson.M{"members": from}).Decode(&sClan); err != nil {
								replyToMsg(messID, errStart+"find_from_clan", peer, bot)
								errl.Println("e: ", err)
								return
							}
						case 1:
							tag := strings.ToUpper(args[4])
							if len(tag) < 3 || len(tag) > 5 {
								replyToMsg(messID, "Некорректный тег", peer, bot)
								return
							}
							if c, err := clans.CountDocuments(ctx, bson.M{"_id": tag}); err != nil {
								replyToMsg(messID, errStart+"count_tag_clan", peer, bot)
								errl.Println("e: ", err)
								return
							} else if c == 0 {
								replyToMsg(messID, "Клана с таким тегом нет...", peer, bot)
								return
							}
							if err := clans.FindOne(ctx, bson.M{"_id": tag}).Decode(&sClan); err != nil {
								replyToMsg(messID, errStart+"find_tag_clan", peer, bot)
								errl.Println("e: ", err)
								return
							}
						default:
							replyToMsg(messID, "СЛИШКОМ. МНОГО. АРГУМЕНТОВ(((", peer, bot)
							return
						}
						var (
							is   bool
							isfr bool
						)
						if is, isfr = isInClattacks(sClan.Tag, clattacks); !is {
							replyToMsg(messID, "Этот клан не учавствует в атаках)", peer, bot)
							return
						}
						var sClat Clattack
						if err := clans.FindOne(ctx, bson.M{
							func() string {
								if isfr {
									return "from"
								}
								return "to"
							}(): sClan.Tag,
						}).Decode(&sClat); err != nil {
							replyToMsg(messID, errStart+"find_clat", peer, bot)
							errl.Println("e: ", err)
							return
						}
						var tocl, frcl Clan
						if err := clans.FindOne(ctx, bson.M{"_id": func() string {
							if isfr {
								return sClat.To
							}
							return sClat.From
						}()}).Decode(func() *Clan {
							if isfr {
								frcl = sClan
								return &tocl
							}
							tocl = sClan
							return &frcl
						}()); err != nil {
							replyToMsg(messID, errStart+"find_sec_clan", peer, bot)
							errl.Println("e: ", err)
							return
						}
						replyToMsg(messID, fmt.Sprintf("От: [%s] %s\nНа: [%s] %s",
							frcl.Tag, frcl.Name,
							tocl.Tag, tocl.Name,
						), peer, bot)
					default:
						replyToMsg(messID, "Что такое "+args[2]+"?", peer, bot)
						return
					}
				case "казна":
					if len(args) == 2 {
						replyToMsg(messID, "жесь", peer, bot)
						return
					}
					switch args[2] {
					case "казна":
						replyToMsg(messID, strings.Repeat("казна ", 42), peer, bot)
						return
					case "снять":
						if len(args) != 4 {
							replyToMsg(messID, "Слишком мало или много аргументов", peer, bot)
							return
						}
						if !isInUsers {
							replyToMsg(messID, "Кланы — приватная территория вомбатов. У тебя вомбата нет.", peer, bot)
							return
						}
						if c, err := clans.CountDocuments(ctx, bson.M{"members": womb.ID}); err != nil {
							replyToMsg(messID, errStart+"count_from_clan", peer, bot)
							errl.Println("e: ", err)
							return
						} else if c == 0 {
							replyToMsg(messID, "Вы не состоите ни в одном клане", peer, bot)
							return
						}
						var sClan Clan
						if err := clans.FindOne(ctx, bson.M{"members": womb.ID}).Decode(&sClan); err != nil {
							replyToMsg(messID, errStart+"find_from_clan", peer, bot)
							errl.Println("e: ", err)
							return
						}
						if !(sClan.Leader == womb.ID || sClan.Banker == womb.ID) {
							replyToMsg(messID, "Ошибка: вы не обладаете правом снимать деньги с казны (только лидер и казначей)",
								peer, bot,
							)
							return
						}
						var take uint64
						if take, err = strconv.ParseUint(args[3], 10, 64); err != nil {
							if args[3] == "всё" {
								take = sClan.Money
							} else {
								replyToMsg(messID,
									"Ошибка: введено не число, либо число больше 2^63, либо отрицательное, либо дробное. короче да.",
									peer, bot,
								)
								return
							}
						}
						if take > sClan.Money {
							replyToMsg(messID, "Запрашиваемая сумма выше количества денег в казне", peer, bot)
							return
						} else if take == 0 {
							replyToMsg(messID, "Хитр(ый/ая) как(ой/ая)", peer, bot)
							return
						}
						if _, err := clans.UpdateOne(ctx, bson.M{"_id": sClan.Tag},
							bson.M{"$inc": bson.M{"money": -int(take)}}); err != nil {
							replyToMsg(messID, errStart+"update_clan", peer, bot)
							errl.Println("e: ", err)
							return
						} else if _, err := users.UpdateOne(ctx, bson.M{"_id": womb.ID},
							bson.M{"$inc": bson.M{"money": int(take)}}); err != nil {
							replyToMsg(messID, errStart+"update_user", peer, bot)
							errl.Println("e: ", err)
							return
						}
						replyToMsg(messID,
							fmt.Sprintf(
								"Вы успешно сняли из казны %d Ш! Теперь в казне %d Ш, а у вас на счету %d",
								take, sClan.Money-take, womb.Money+take,
							),
							peer, bot,
						)
					case "положить":
						if len(args) != 4 {
							replyToMsg(messID, "Слишком много или мало аргументов", peer, bot)
							return
						}
						if !isInUsers {
							replyToMsg(messID, "Кланы — приватная территория вомбатов. У тебя вомбата нет.", peer, bot)
							return
						}
						if c, err := clans.CountDocuments(ctx, bson.M{"members": from}); err != nil {
							replyToMsg(messID, errStart+"count_from_clan", peer, bot)
							errl.Println("e: ", err)
							return
						} else if c == 0 {
							replyToMsg(messID, "Вы не состоите ни в одном клане", peer, bot)
							return
						}
						var sClan Clan
						if err := clans.FindOne(ctx, bson.M{"members": from}).Decode(&sClan); err != nil {
							replyToMsg(messID, errStart+"find_from_clan", peer, bot)
							errl.Println("e: ", err)
							return
						}
						var (
							take uint64
							err  error
						)
						if take, err = strconv.ParseUint(args[3], 10, 64); err != nil {
							replyToMsg(messID,
								"Ошибка: введено не число, либо число больше 2^63, либо отрицательное, либо дробное. короче да.",
								peer, bot,
							)
							return
						} else if take > womb.Money {
							replyToMsg(messID, "Сумма, которую вы хотите положить, больше кол-ва денег на вашем счету", peer, bot)
							return
						} else if take == 0 {
							replyToMsg(messID, "блин", peer, bot)
							return
						}
						if _, err := users.UpdateOne(ctx, bson.M{"_id": womb.ID}, bson.M{
							"$inc": bson.M{
								"money": -int(take),
							},
						}); err != nil {
							replyToMsg(messID, errStart+"update_user", peer, bot)
							errl.Println("e: ", err)
							return
						} else if _, err := clans.UpdateOne(ctx, bson.M{"_id": sClan.Tag}, bson.M{
							"$inc": bson.M{
								"money": int(take),
							},
						}); err != nil {
							replyToMsg(messID, errStart+"update_user", peer, bot)
							errl.Println("e: ", err)
							return
						}
						replyToMsg(messID,
							fmt.Sprintf(
								"Вы положили %d Ш в казну. Теперь в казне %d Ш, а у вас %d",
								take, sClan.Money+take, womb.Money-take,
							),
							peer, bot,
						)
					default:
						replyToMsg(messID, fmt.Sprintf("Что такое `%s`?", args[2]),
							peer, bot,
						)
					}
				default:
					replyToMsg(messID, fmt.Sprintf("Что такое `%s`?", args[1]),
						peer, bot,
					)
				}
			} else if args := strings.Fields(txt); len(args) >= 3 && strings.ToLower(args[0]) == "sendmsg" {
				if !hasTitle(0, womb.Titles) {
					return
				}
				to, err := strconv.Atoi(args[1])
				if err != nil {
					replyToMsg(messID, errStart+"sendmsg: atoi", peer, bot)
					errl.Println("e: ", err)
					return
				}
				sendMsgMD(strings.Join(args[2:], " "), int64(to), bot)
				replyToMsg(messID, "Запрос отправлен успешно!", peer, bot)
			}
		}(update, titles, bot, users, titlesC, attacks, imgsC, bank, clans, clattacks)
	}
}
