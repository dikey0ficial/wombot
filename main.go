package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/go-vk-api/vk"
	lp "github.com/go-vk-api/vk/longpoll/user"
	jsoniter "github.com/json-iterator/go"
)

type Config struct {
	Token string `json:"token"`
}

type User struct {
	Name     string `json:"name"`
	XP       uint32 `json:"xp"`
	Health   uint32 `json:"health"`
	Force    uint32 `json:"force"`
	Money    uint64 `json:"money"`
	LastTime string `json:"last_time"` //when was last reward got
}

var users map[int64]User = map[int64]User{}

var standartNicknames []string = []string{"Вомбатыч", "Вомбатус", "wombatkiller2007", "wombatik", "батвом", "Табмов"}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func checkerr(err error) {
	if err != nil && err.Error() != "EOF" {
		log.Println("ERROR\n\n", err)
	}
}

func loadConfig() (Config, error) {
	file, err := os.Open("config.json")
	if err != nil {
		return Config{}, err
	}
	defer file.Close()
	result := Config{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&result)
	file.Close()
	return result, err
}

func loadUsers() {
	file, err := os.Open("users.json")
	checkerr(err)
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&users)
	file.Close()
	checkerr(err)
}

func saveUsers() {
	jsonFile, err := os.Create("users.json")
	checkerr(err)
	data, err := json.Marshal(users)
	checkerr(err)
	defer jsonFile.Close()
	jsonFile.Write(data)
	jsonFile.Close()
}

func isInList(str string, list []string) bool {
	for _, elem := range list {
		if strings.ToLower(str) == strings.ToLower(elem) {
			return true
		}
	}
	return false
}

func sendMsg(message string, peer int64, client *vk.Client) error {
	rand.Seed(time.Now().UnixNano())
	err := client.CallMethod("messages.send", vk.RequestParams{
		"peer_id":   peer,
		"message":   message,
		"random_id": rand.Int63(),
	}, nil)
	return err
}

func main() {
	conf := Config{}
	conf, err := loadConfig()
	checkerr(err)

	loadUsers()

	client, err := vk.NewClientWithOptions(
		vk.WithToken(conf.Token),
	)
	checkerr(err)

	longpoll, err := lp.NewWithOptions(client, lp.WithMode(lp.ReceiveAttachments))
	checkerr(err)
	stream, err := longpoll.GetUpdatesStream(0)
	checkerr(err)

	log.Println("Start!")

	for update := range stream.Updates {
		switch data := update.Data.(type) {
		case *lp.NewMessage:
			if data.PeerID == -201237807 || data.PeerID == 2000000001 {
				break
			}
			peer, txt := data.PeerID, data.Text
			womb, isInUsers := users[peer]

			log.Println(peer, womb.Name, txt)

			if isInList(txt, []string{"старт", "начать", "/старт", "/start", "start"}) {
				if isInUsers {
					sendMsg(fmt.Sprintf("Здравствуйте, %s!", users[peer].Name), peer, client)
				} else {
					sendMsg("Драсьте. Возьмите вомбата с командой `Взять вомбата` или `Купить вомбата у арабов`. напиши `помощь`", peer, client)
				}
			} else if isInList(txt, []string{"взять вомбата", "купить вомбата у арабов", "хочу вомбата"}) {
				if isInUsers {
					sendMsg("У тебя как бы уже есть вомбат лолкек. Если хочешь от него избавиться, то напиши `приготовить шашлык`", peer, client)
				} else {
					users[peer] = User{
						Name:   standartNicknames[rand.Intn(len(standartNicknames))],
						XP:     0,
						Health: 5,
						Force:  2,
						Money:  10,
					}
					womb = users[peer]
					saveUsers()
					sendMsg(fmt.Sprintf("Поздравляю, у тебя появился вомбат! Ему выдалось имя `%s`. Ты можешь поменять имя командой `Поменять имя [имя]` за 3 монеты", womb.Name), peer, client)
				}
			} else if isInList(txt, []string{"приготовить шашлык", "продать вомбата арабам", "слить вомбата в унитаз"}) {
				if isInUsers {
					delete(users, peer)
					saveUsers()
					sendMsg("Вы уничтожили вомбата в количестве 1 штука. Вы - нехорошее существо", peer, client)
				} else {
					sendMsg("Но у вас нет вомбата...", peer, client)
				}
			} else if isInList(txt, []string{"о вомбате", "вомбат инфо"}) {
				if isInUsers {
					sendMsg(fmt.Sprintf("Вомбат по имени %s имеет: \n 🕳 %d XP \n ❤ %d здоровья \n ⚡ %d мощи \n 💰 %d шишей", womb.Name, womb.XP, womb.Health, womb.Force, womb.Money), peer, client)
				} else {
					sendMsg("У вас ещё нет вомбата...", peer, client)
				}
			} else if strings.HasPrefix(strings.ToLower(txt), "поменять имя") {
				if isInUsers {
					name := strings.Title(strings.TrimSpace(strings.TrimPrefix(strings.ToLower(txt), "поменять имя ")))
					if womb.Money >= 3 {
						if isInList(name, []string{"admin", "вoмбoт", "вoмбoт", "вомбoт", "вомбот"}) {
							sendMsg("Такие никнеймы заводить нельзя", peer, client)
						} else if name != "" {
							womb.Money -= 3
							womb.Name = name
							users[peer] = womb
							saveUsers()
							sendMsg(fmt.Sprintf("Теперь вашего вомбата зовут %s. С вашего счёта сняли 3 шиша", name), peer, client)
						} else {
							sendMsg("У вас пустое имя...", peer, client)
						}
					} else {
						sendMsg("Мало шишей блин нафиг!!!!", peer, client)
					}
				} else {
					sendMsg("Да блин нафиг, вы вобмата забыли завести!!!!!!!", peer, client)
				}
			} else if isInList(txt, []string{"помощь", "хелп", "help", "команды"}) {
				sendMsg("vk.com/@wombat_bot-help", peer, client)
			} else if isInList(txt, []string{"купить здоровье", "прокачка здоровья", "прокачать здоровье"}) {
				if isInUsers {
					if womb.Money >= 5 {
						womb.Money -= 5
						womb.Health++
						users[peer] = womb
						saveUsers()
						sendMsg(fmt.Sprintf("Поздравляю! Теперь у вас %d здоровья и %d шишей", womb.Health, womb.Money), peer, client)
					} else {
						sendMsg("Надо накопить побольше шишей! 1 здоровье = 5 шишей", peer, client)
					}
				} else {
					sendMsg("У тя ваще вобата нет...", peer, client)
				}
			} else if isInList(txt, []string{"купить мощь", "прокачка мощи", "прокачка силы", "прокачать мощь", "прокачать силу"}) {
				if isInUsers {
					if womb.Money >= 3 {
						womb.Money -= 3
						womb.Force++
						users[peer] = womb
						saveUsers()
						sendMsg(fmt.Sprintf("Поздравляю! Теперь у вас %d мощи и %d шишей", womb.Force, womb.Money), peer, client)
					} else {
						sendMsg("Надо накопить побольше шишей! 1 мощь = 3 шиша", peer, client)
					}
				} else {
					sendMsg("У тя ваще вобата нет...", peer, client)
				}
			} else if isInList(txt, []string{"поиск денег"}) {
				if isInUsers {
					if womb.Money >= 1 {
						womb.Money--
						rand.Seed(time.Now().UnixNano())
						if rand.Int()%2 == 0 {
							rand.Seed(time.Now().UnixNano())
							win := rand.Intn(9) + 1
							womb.Money += uint64(win)
							sendMsg(fmt.Sprintf("Поздравляем! Вы нашли на дороге %d шишей! Теперь их у вас %d", win, womb.Money), peer, client)
						} else {
							sendMsg("Вы заплатили один шиш охранникам денежной дорожки, но увы, вы так ничего и не нашли", peer, client)
						}
						users[peer] = womb
						saveUsers()
					} else {
						sendMsg("Охранники тебя прогнали; они требуют шиш за проход, а у тебя и шиша-то нет", peer, client)
					}
				} else {
					sendMsg("А ты куда? У тебя вомбата нет...", peer, client)
				}
			}
		}
	}
}
