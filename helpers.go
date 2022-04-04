package main

import (
	"context"
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"math/rand"
	"strings"
	"time"
)

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
func sendMsg(message string, chatID int64, bot *tg.BotAPI) (int, error) {
	msg := tg.NewMessage(chatID, message)
	mess, err := bot.Send(msg)

	return mess.MessageID, err
}

// sendMsgMD отправляет сообщение с markdown
func sendMsgMD(message string, chatID int64, bot *tg.BotAPI) int {
	msg := tg.NewMessage(chatID, message)
	msg.ParseMode = "markdown"
	mess, err := bot.Send(msg)

	if err != nil {
		debl.Println(chatID)
	}
	return mess.MessageID
}

// replyToMsg отвечает обычным сообщением
func replyToMsg(replyID int, message string, chatID int64, bot *tg.BotAPI) (int, error) {
	msg := tg.NewMessage(chatID, message)
	msg.ReplyToMessageID = replyID
	mess, err := bot.Send(msg)

	return mess.MessageID, err
}

// replyToMsgMDNL отвечает сообщением с markdown без ссылок
func replyToMsgMDNL(replyID int, message string, chatID int64, bot *tg.BotAPI) (int, error) {
	msg := tg.NewMessage(chatID, message)
	msg.ReplyToMessageID = replyID
	msg.ParseMode = "markdown"
	msg.DisableWebPagePreview = true
	mess, err := bot.Send(msg)

	return mess.MessageID, err
}

// sendPhoto отправляет текст с картинкой
func sendPhoto(id, caption string, chatID int64, bot *tg.BotAPI) (int, error) {
	msg := tg.NewPhoto(chatID, tg.FileID(id))
	msg.Caption = caption
	mess, err := bot.Send(msg)

	return mess.MessageID, err
}

// sendPhotoMD отправляет текст с markdown с картинкой
func sendPhotoMD(id, caption string, chatID int64, bot *tg.BotAPI) int {
	msg := tg.NewPhoto(chatID, tg.FileID(id))
	msg.Caption = caption
	msg.ParseMode = "markdown"
	mess, err := bot.Send(msg)

	if err != nil {
		debl.Println(chatID)
	}
	return mess.MessageID
}

// replyToMsgMD отвечает сообщением с markdown
func replyToMsgMD(replyID int, message string, chatID int64, bot *tg.BotAPI) (int, error) {
	msg := tg.NewMessage(chatID, message)
	msg.ReplyToMessageID = replyID
	msg.ParseMode = "markdown"
	mess, err := bot.Send(msg)

	return mess.MessageID, err
}

// replyWithPhotoMD отвечает картинкой с текстом с markdown
func replyWithPhotoMD(replyID int, id, caption string, chatID int64, bot *tg.BotAPI) (int, error) {
	msg := tg.NewPhoto(chatID, tg.FileID(id))
	msg.Caption = caption
	msg.ReplyToMessageID = replyID
	msg.ParseMode = "markdown"
	mess, err := bot.Send(msg)

	return mess.MessageID, err
}

// replyWithPhotoMD отвечает картинкой с текстом
func replyWithPhoto(replyID int, id, caption string, chatID int64, bot *tg.BotAPI) (int, error) {
	msg := tg.NewPhoto(chatID, tg.FileID(id))
	msg.Caption = caption
	msg.ReplyToMessageID = replyID
	mess, err := bot.Send(msg)

	return mess.MessageID, err
}

// isInAttacks возвращает информацию, есть ли существо в атаках и
// отправитель ли он
func isInAttacks(id int64, attacks *mongo.Collection) (isIn, isFrom bool) {
	if f, err := attacks.CountDocuments(ctx, bson.M{"from": id}); f > 0 && err == nil {
		isFrom = true
	} else if err != nil {

	}
	var isTo bool
	if t, err := attacks.CountDocuments(ctx, bson.M{"to": id}); t > 0 && err == nil {
		isTo = true
	} else if err != nil {

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

	}
	var isTo bool
	if t, err := attacks.CountDocuments(ctx, bson.M{"to": id}); t > 0 && err == nil {
		isTo = true
	} else if err != nil {

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
func delMsg(ID int, chatID int64, bot *tg.BotAPI) error {
	delConfig := tg.DeleteMessageConfig{
		ChatID:    chatID,
		MessageID: ID,
	}
	_, err := bot.Request(delConfig)
	return err
}

func editMsg(mid int, txt string, chatID int64, bot *tg.BotAPI) error {
	editConfig := tg.EditMessageTextConfig{
		BaseEdit: tg.BaseEdit{
			ChatID:    chatID,
			MessageID: mid,
		},
		Text: txt,
	}
	_, err := bot.Request(editConfig)
	return err
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

func cins(s string) primitive.Regex {
	return primitive.Regex{
		Pattern: fmt.Sprintf("^%s$", s),
		Options: "i",
	}
}

func b2s(b bool) string {
	if b {
		return "да"
	}
	return "нет"
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

func isGroup(m *tg.Message) bool {
	if m == nil || m.Chat == nil {
		return false
	}
	return m.Chat.IsGroup() || m.Chat.IsSuperGroup()
}

func getIsInUsers(id int64) (bool, error) {
	rCount, err := users.CountDocuments(ctx, bson.M{"_id": id})
	if err != nil {
		return false, err
	}
	return rCount != 0, nil
}

func getIsBanked(id int64) (bool, error) {
	rCount, err := bank.CountDocuments(ctx, bson.M{"_id": id})
	return rCount != 0, err
}

func wombFilter(womb User) bson.M {
	return bson.M{"_id": womb.ID}
}

func randomString(arr ...string) string {
	if len(arr) == 0 {
		return ""
	}
	return arr[rand.Intn(len(arr))]
}
