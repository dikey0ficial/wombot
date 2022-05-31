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
	infl, messl, errl, debl *log.Logger
	ctx                     = context.Background()
	conf                    = struct {
		Token    string `toml:"tg_token" env:"TGTOKEN"`
		MongoURL string `toml:"mongo_url" env:"MONGOURL"`
		// 0 — no information about messages
		// 1 — about every command
		// 2 — about every message (text)
		LogLevel  uint8 `toml:"log_level" env:"LOGLVL"`
		SupChatID int64 `toml:"support_chat_id" env:"SUPCHATID"`
	}{}
	bot Bot

	mongoClient                            *mongo.Client
	db                                     *mongo.Database
	users, attacks, bank, clans, clattacks *mongo.Collection
	titlesC, imgsC                         *mongo.Collection
	titles                                 []Title
)

func init() {
	infl = log.New(os.Stdout, "[  INF  ]\t", log.Ltime)
	messl = log.New(os.Stdout, "[  MSG  ]\t", log.Ltime)
	errl = log.New(os.Stderr, "[ ERROR ]\t", log.Ltime)
	debl = log.New(os.Stdout, "[ DEBUG ]\t", log.Ltime|log.Lshortfile)
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
	var err error
	bot.BotAPI, err = tg.NewBotAPI(conf.Token)
	if err != nil {
		panic(err)
	}

	u := tg.NewUpdate(0)
	u.Timeout = 1
	updates := bot.GetUpdatesChan(u)
	if err != nil {
		panic(err)
	}
	var wg = sync.WaitGroup{}

	defer func() {
		wg.Wait()
		infl.Println("==end==")
	}()

	infl.Println("==start==")

	for update := range updates {
		wg.Add(1)
		go func(update tg.Update) {
			defer wg.Done()
			var (
				cmdName string   = "-"
				args    []string = make([]string, 0)
				womb    User     = User{}
				messID  int      = 0
			)

			defer func() {
				if e := recover(); e != nil {
					errl.Printf("Goroutine failed (%s): %v\n", cmdName, e)
				}
			}()

			if update.Message != nil {
				args = strings.Fields(update.Message.Text)
				messID = update.Message.MessageID
				_ = users.FindOne(ctx, bson.M{"_id": update.Message.From.ID}).Decode(&womb)

				if conf.LogLevel == 2 {
					logMessage(*update.Message)
				}
			}

			for _, cmd := range commands {
				if cmd.Is(args, update) {
					if conf.LogLevel == 1 {
						logMessage(*update.Message)
					}

					cmdName := cmd.Name
					err := cmd.Action(args, update, womb)
					if err != nil {
						errl.Printf("%d: %s: %v\n", messID, cmdName, err)
						bot.ReplyWithMessage(
							update.Message.MessageID,
							"Произошла ошибка... ответьте на это сообщение командой /admin",
							update.Message.Chat.ID,
						)
					}
					break
				}
			}
		}(update)
	}

}
