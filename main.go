package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	// "sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-vk-api/vk"
	lp "github.com/go-vk-api/vk/longpoll/user"
	jsoniter "github.com/json-iterator/go"
)

// Config нужен для токена и мб настроек
type Config struct {
	Token string `json:"token"`
}

// Title — описание титула
type Title struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
}

// User — описание пользователя
type User struct { // параметры юзера
	Name   string           `json:"name"`
	XP     uint32           `json:"xp"`
	Health uint32           `json:"health"`
	Force  uint32           `json:"force"`
	Money  uint64           `json:"money"`
	Titles map[uint16]Title `json:"titles"`
	Subs   map[string]int64 `json:"subs"`
}

/*
// Sorter нужен для сортировки
type Sorter interface {
	Len() int
	Less(i, j int) bool
	Swap(i, j int)
}

// ByMoney ;
type ByMoney map[int64]User

//Len ;
func (a ByMoney) Len() int { return len(a) }

//Less ;
func (a ByMoney) Less(i, j int) bool { return a[int64(i)].Money < a[int64(j)].Money }

//Swap ;
func (a ByMoney) Swap(i, j int) { a[int64(i)], a[int64(j)] = a[int64(j)], a[int64(i)] }

// ByXP ;
type ByXP map[int64]User

//Len ;
func (a ByXP) Len() int { return len(a) }

//Less ;
func (a ByXP) Less(i, j int) bool { return a[int64(i)].XP < a[int64(j)].XP }

//Swap ;
func (a ByXP) Swap(i, j int) { a[int64(i)], a[int64(j)] = a[int64(j)], a[int64(i)] }

// ByForce ;
type ByForce map[int64]User

//Len ;
func (a ByForce) Len() int { return len(a) }

//Less ;
func (a ByForce) Less(i, j int) bool { return a[int64(i)].Force < a[int64(j)].Force }

//Swap ;
func (a ByForce) Swap(i, j int) { a[int64(i)], a[int64(j)] = a[int64(j)], a[int64(i)] }

*/

var users = map[int64]User{}

var titles = map[uint16]Title{
	0: Title{
		Name: "Вомботестер",
		Desc: "Тестирует вомбота; даёт право пользоваться devtools",
	},
	1: Title{
		Name: "Спамер",
		Desc: "Настолько достал админа, что получил этот титул; забирает право удалить вомбата",
	},
}

var standartNicknames []string = []string{"Вомбатыч", "Вомбатус", "wombatkiller2007", "wombatik", "батвом", "Табмов", "Вомбабушка"}

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

func sendMsg(message string, peer int64, client *vk.Client) {
	rand.Seed(time.Now().UnixNano())
	err := client.CallMethod("messages.send", vk.RequestParams{
		"peer_id":   peer,
		"message":   message,
		"random_id": rand.Int63(),
	}, nil)
	checkerr(err)
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

			if isInList(txt, []string{"старт", "начать", "/старт", "/start", "start", "привет"}) {
				if isInUsers {
					sendMsg(fmt.Sprintf("Здравствуйте, %s!", users[peer].Name), peer, client)
				} else {
					sendMsg("Привет! Для того, чтобы ознакомиться с механикой бота, почитай справку https://vk.com/@wombat_bot-help (она также доступна по команде `помощь`. Чтобы завести вомбата, напиши `взять вомбата`. Приятной игры!", peer, client)
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
						Titles: map[uint16]Title{},
						Subs:   map[string]int64{},
					}
					womb = users[peer]
					saveUsers()
					sendMsg(fmt.Sprintf("Поздравляю, у тебя появился вомбат! Ему выдалось имя `%s`. Ты можешь поменять имя командой `Поменять имя [имя]` за 3 монеты", womb.Name), peer, client)
				}
			} else if strings.HasPrefix(strings.ToLower(txt), "devtools") {
				if _, ok := womb.Titles[0]; ok {
					cmd := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(txt), "devtools"))
					if strings.HasPrefix(cmd, "set money") {
						strNewMoney := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(cmd), "set money "))
						if _, err := strconv.ParseInt(strNewMoney, 10, 64); err == nil {
							i, err := strconv.Atoi(strNewMoney)
							checkerr(err)
							womb.Money = uint64(i)
							users[peer] = womb
							saveUsers()
							sendMsg(fmt.Sprintf("Операция проведена успешно! Шишей на счету: %d", womb.Money), peer, client)
						} else {
							sendMsg("Ошибка: неправильный синтаксис. Синтаксис команды: `devtools set money {кол-во шишей}`", peer, client)
						}
					} else if cmd == "help" {
						sendMsg("https://vk.com/@wombat_bot-kak-polzovatsya-devtools", peer, client)
					}
				} else if strings.TrimSpace(txt) == "devtools on" {
					womb.Titles[0] = titles[0]
					users[peer] = womb
					saveUsers()
					sendMsg("Выдан титул \"Вомботестер\" (ID: 0)", peer, client)
				}
			} else if isInList(txt, []string{"приготовить шашлык", "продать вомбата арабам", "слить вомбата в унитаз"}) {
				if isInUsers {
					if _, ok := womb.Titles[1]; !ok {
						delete(users, peer)
						saveUsers()
						sendMsg("Вы уничтожили вомбата в количестве 1 штука. Вы - нехорошее существо", peer, client)
					} else {
						sendMsg("Ошибка: вы лишены права уничтожать вомбата; обратитксь к @dikey_oficial за разрешением", peer, client)
					}
				} else {
					sendMsg("Но у вас нет вомбата...", peer, client)
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
				sendMsg("https://vk.com/@wombat_bot-help", peer, client)
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
						if ch := rand.Int(); ch%2 == 0 {
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
			} else if strings.HasPrefix(strings.ToLower(txt), "о титуле") {
				strID := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(txt), "о титуле"))
				if strID == "" {
					sendMsg("Ошибка: пустой ID титула", peer, client)
				} else if _, err := strconv.ParseInt(strID, 10, 64); err == nil {
					i, err := strconv.Atoi(strID)
					checkerr(err)
					ID := uint16(i)
					if elem, ok := titles[ID]; ok {
						sendMsg(fmt.Sprintf("%s | ID: %d\n%s", elem.Name, ID, elem.Desc), peer, client)
					} else {
						sendMsg(fmt.Sprintf("Ошибка: не найдено титула по ID %d", ID), peer, client)
					}
				} else {
					sendMsg("Ошибка: неправильный синтаксис. Синтаксис команды: `о титуле {ID титула}`", peer, client)
				}
				// } else if strings.HasPrefix(strings.ToLower(txt), "рейтинг") {
				// 	sorting := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(txt), "рейтинг"))
				// 	sortedUsers := users
				// 	if isInList(sorting, []string{"шиши", "шиш", "деньги", "монеты", "монетки"}) {
				// 		sort.Sort(ByMoney(sortedUsers))

				// 		}
				// 	}
			} else if strings.HasPrefix(strings.ToLower(txt), "подписаться") {
				args := strings.Fields(strings.TrimSpace(strings.TrimPrefix(strings.ToLower(txt), "подписаться")))
				if len(args) == 0 {
					sendMsg("Ошибка: пропущены аргументы `ID` и `алиас`. Синтаксис команды: `подписаться [ID] [алиас (без пробелов)]`", peer, client)
				} else if len(args) == 1 {
					sendMsg("Ошибка: пропущен аргумент `алиас`. Синтаксис команды: `подписаться [ID] [алиас (без пробелов)]`", peer, client)
				} else if len(args) == 2 {
					if ID, err := strconv.ParseInt(args[0], 10, 64); err == nil {
						if _, err := strconv.ParseInt(args[1], 10, 64); err == nil {
							sendMsg("Ошибка: алиас не должен быть числом", peer, client)
						} else {
							if elem, ok := womb.Subs[args[1]]; !ok {
								if _, ok := users[ID]; ok {
									womb.Subs[args[1]] = ID
									users[peer] = womb
									saveUsers()
									sendMsg(fmt.Sprintf("Вомбат с ID %d добавлен в ваши подписки", ID), peer, client)
								} else {
									sendMsg(fmt.Sprintf("Ошибка: пользователя с ID %d не найдено", ID), peer, client)
								}
							} else {
								sendMsg(fmt.Sprintf("Ошибка: алиас %s занят id %d", args[1], elem), peer, client)
							}
						}
					} else {
						sendMsg(fmt.Sprintf("Ошибка: `%s` не является числом", args[0]), peer, client)
					}
				} else {
					sendMsg("Ошибка: слишком много аргументов. Синтаксис команды: `подписаться [ID] [алиас (без пробелов)]", peer, client)
				}
			} else if strings.HasPrefix(strings.ToLower(txt), "отписаться") {
				alias := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(txt), "отписаться"))
				if _, ok := womb.Subs[alias]; ok {
					delete(womb.Subs, alias)
					users[peer] = womb
					saveUsers()
					sendMsg(fmt.Sprintf("Вы отписались от пользователя с алиасом %s", alias), peer, client)
				} else {
					sendMsg(fmt.Sprintf("Ошибка: вы не подписаны на пользователя с алиасом `%s`", alias), peer, client)
				}
			} else if isInList(txt, []string{"подписки", "мои подписки", "список подписок"}) {
				strSubs := "Вот список твоих подписок:"
				if len(womb.Subs) != 0 {
					for alias, id := range womb.Subs {
						if tWomb, ok := users[id]; ok {
							strSubs += fmt.Sprintf("\n %s | ID: %d | Алиас: %s", tWomb.Name, id, alias)
						} else {
							strSubs += fmt.Sprintf("\n Ошибка: пользователь по алиасу `%s` не найден", alias)
						}
					}
				} else {
					strSubs = "У тебя пока ещё нет подписок"
				}
				sendMsg(strSubs, peer, client)
			} else if isInList(txt, []string{"мои вомбаты", "мои вомбатроны", "вомбатроны", "лента подписок"}) {
				if len(womb.Subs) == 0 {
					sendMsg("У тебя пока ещё нет подписок", peer, client)
					continue
				}
				for alias, ID := range womb.Subs {
					if tWomb, ok := users[ID]; ok {
						strTitles := ""
						for id, elem := range tWomb.Titles {
							strTitles += fmt.Sprintf("%s (ID: %d) | ", elem.Name, id)
						}
						strTitles = strings.TrimSuffix(strTitles, " | ")
						if strings.TrimSpace(strTitles) == "" {
							strTitles = "нет"
						}
						sendMsg(fmt.Sprintf("Вомбат  %s (ID: %d)\nТитулы: %s\n 🕳 %d XP \n ❤ %d здоровья \n ⚡ %d мощи \n 💰 %d шишей", tWomb.Name, ID, strTitles, tWomb.XP, tWomb.Health, tWomb.Force, tWomb.Money), peer, client)
					} else {
						sendMsg(fmt.Sprintf("Ошибка: подписчика с алиасом `%s` не обнаружено", alias), peer, client)
					}
				}
			} else if strings.HasPrefix(strings.ToLower(txt), "о вомбате") {
				strID := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(txt), "о вомбате"))
				log.Println(strID, txt)
				if strID == "" {
					if isInUsers {
						strTitles := ""
						for id, elem := range womb.Titles {
							strTitles += fmt.Sprintf("%s (ID: %d) | ", elem.Name, id)
						}
						strTitles = strings.TrimSuffix(strTitles, " | ")
						if strings.TrimSpace(strTitles) == "" {
							strTitles = "нет"
						}
						sendMsg(fmt.Sprintf("Вомбат  %s (ID: %d)\nТитулы: %s\n 🕳 %d XP \n ❤ %d здоровья \n ⚡ %d мощи \n 💰 %d шишей", womb.Name, peer, strTitles, womb.XP, womb.Health, womb.Force, womb.Money), peer, client)
					} else {
						sendMsg("У вас ещё нет вомбата...", peer, client)
					}
				} else if _, err := strconv.ParseInt(strID, 10, 64); err == nil {
					ID, err := strconv.ParseInt(strID, 10, 64)
					checkerr(err)
					if _, ok := users[ID]; ok {
						tWomb := users[ID]
						strTitles := ""
						for id, elem := range tWomb.Titles {
							strTitles += fmt.Sprintf("%s (ID: %d) | ", elem.Name, id)
						}
						strTitles = strings.TrimSuffix(strTitles, " | ")
						if strings.TrimSpace(strTitles) == "" {
							strTitles = "нет"
						}
						sendMsg(fmt.Sprintf("Вомбат  %s (ID: %d)\nТитулы: %s\n 🕳 %d XP \n ❤ %d здоровья \n ⚡ %d мощи \n 💰 %d шишей", tWomb.Name, ID, strTitles, tWomb.XP, tWomb.Health, tWomb.Force, tWomb.Money), peer, client)
					} else {
						sendMsg(fmt.Sprintf("Ошибка: игрока с ID %d не найдено", ID), peer, client)
					}
				} else if _, ok := womb.Subs[strID]; ok {
					if tWomb, ok := users[womb.Subs[strID]]; ok {
						strTitles := ""
						for id, elem := range tWomb.Titles {
							strTitles += fmt.Sprintf("%s (ID: %d) | ", elem.Name, id)
						}
						strTitles = strings.TrimSuffix(strTitles, " | ")
						if strings.TrimSpace(strTitles) == "" {
							strTitles = "нет"
						}
						sendMsg(fmt.Sprintf("Вомбат  %s (ID: %d)\nТитулы: %s\n 🕳 %d XP \n ❤ %d здоровья \n ⚡ %d мощи \n 💰 %d шишей", tWomb.Name, womb.Subs[strID], strTitles, tWomb.XP, tWomb.Health, tWomb.Force, tWomb.Money), peer, client)
					} else {
						sendMsg(fmt.Sprintf("Ошибка: неправильный алиас `%s` или не найден пользователь с ID %d. Обратитесь к @dikey_oficial, если такой вомбат есть", strID, womb.Subs[strID]), peer, client)
					}
				} else {
					sendMsg(fmt.Sprintf("Ошибка: не найден алиас `%s`", strID), peer, client)
				}
			}
		}
	}
}
