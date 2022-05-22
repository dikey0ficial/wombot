package main

import (
	"context"
	"github.com/BurntSushi/toml"
	"github.com/caarlos0/env"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"

	"io/ioutil"
	"log"
	"os"
	"strings"
)

var (
	infl, errl, debl, servl *log.Logger
	ctx                     = context.Background()
	conf                    = struct {
		Token     string `toml:"tg_token" env:"TGTOKEN"`
		MongoURL  string `toml:"mongo_url" env:"MONGOURL"`
		SupChatID int64  `toml:"support_chat_id" env:"SUPCHATID"`
	}{}
	bot *tg.BotAPI

	mongoClient                            *mongo.Client
	db                                     *mongo.Database
	users, attacks, bank, clans, clattacks *mongo.Collection
	titlesC, imgsC                         *mongo.Collection
	titles                                 []Title
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
	mongoClient, err := mongo.NewClient(options.Client().ApplyURI(conf.MongoURL))
	if err != nil {
		errl.Println(err)
		os.Exit(1)
	}
	err = mongoClient.Connect(ctx)
	if err != nil {
		errl.Println(err)
		os.Exit(1)
	}
	db = mongoClient.Database("wombot")

	users = db.Collection("users")

	attacks = db.Collection("attacks")

	bank = db.Collection("bank")

	clans = db.Collection("clans")
	clattacks = db.Collection("clattacks")

	titlesC = db.Collection("titles")
	cur, err := titlesC.Find(ctx, bson.M{})
	if err != nil {
		errl.Println(err)
		os.Exit(1)
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var nextOne Title
		err := cur.Decode(&nextOne)
		if err != nil {
			errl.Printf("Lost title because error: %v\n", err)
			continue
		}
		titles = append(titles, nextOne)
	}
	imgsC = db.Collection("imgs")
}

func main() {
	// init
	bot, err := tg.NewBotAPI(conf.Token)
	checkerr(err)

	u := tg.NewUpdate(0)
	u.Timeout = 1
	updates := bot.GetUpdatesChan(u)
	checkerr(err)
	var wg = sync.WaitGroup{}

	defer func() {
		wg.Wait()
		infl.Println("==end==")
	}()

	infl.Println("==start==")

	for update := range updates {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var cmdName string = "-"
			defer func() {
				if e := recover(); e != nil {
					errl.Printf("Goroutine failed (%s): %v\n", cmdName, e)
				}
			}()
			var args = make([]string, 0)
			var womb = User{}
			debl.Println("!")
			if update.Message != nil {
				args = strings.Fields(update.Message.Text)
				// ignoring error because yes
				_ = users.FindOne(ctx, bson.M{"_id": update.Message.From.ID}).Decode(&womb)
			}
			debl.Println("!")
			for _, cmd := range commands {
				debl.Println(cmd.Name)
				if cmd.Is(args, update) {
					cmdName := cmd.Name
					debl.Println(cmdName)
					err := cmd.Action(args, update, womb)
					if err != nil {
						errl.Printf("%s: %v", cmdName, err)
						_, err = replyToMsg(
							update.Message.MessageID,
							"Произошла ошибка... ответьте на это сообщение командой /admin",
							update.Message.Chat.ID, bot,
						)
					}
					break
				}
			}
		}()
	}

}
