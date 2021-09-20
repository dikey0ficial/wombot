package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	// "sort"
	"github.com/go-vk-api/vk"
	lp "github.com/go-vk-api/vk/longpoll/user"
	jsoniter "github.com/json-iterator/go"
	db "github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/mongo"
	"strconv"
	"strings"
	"time"
)

// MongoConfig нужен для настроек монги
type MongoConfig struct {
	Database string `json:"db"`
	Host     string `json:"host"`
	Login    string `json:"login"`
	Password string `json:"pwrd"`
}

// Config нужен для настроек
type Config struct {
	Token string      `json:"token"`
	Mongo MongoConfig `json:"mongo"`
}

// Title — описание титула
type Title struct {
	Name string `db:"name"`
	Desc string `db:"desc"`
}

// User — описание пользователя
type User struct { // параметры юзера
	ID     int64            `db:"_id"`
	Name   string           `db:"name"`
	XP     uint32           `db:"xp"`
	Health uint32           `db:"health"`
	Force  uint32           `db:"force"`
	Money  uint64           `db:"money"`
	Titles map[uint16]Title `db:"titles"`
	Subs   map[string]int64 `db:"subs"`
}

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

func initSess() db.Session {
	conf, err := loadConfig()
	checkerr(err)
	sess, err := mongo.Open(mongo.ConnectionURL{
		Database: conf.Mongo.Database,
		Host:     conf.Mongo.Host,
		User:     conf.Mongo.Login,
		Password: conf.Mongo.Password,
	})
	checkerr(err)
	return sess
}

var sess = initSess()

var standartNicknames []string = []string{"Вомбатыч", "Вомбатус", "wombatkiller2007", "wombatik", "батвом", "Табмов", "Вомбабушка"}

var users db.Collection

func loadUsers() {
	users = sess.Collection("users")
}

var titles db.Collection

func loadTitles() {
	titles = sess.Collection("titles")
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
	defer sess.Close()
	conf, err := loadConfig()
	checkerr(err)

	loadUsers()
	loadTitles()

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

			womb := User{}
			wombRes := users.Find(fmt.Sprintf(`{"_id":%d}}`, peer))

			rCount, err := wombRes.Count()
			checkerr(err)
			isInUsers := rCount != 0
			if isInUsers {
				err = wombRes.One(&womb)
				checkerr(err)
			}

			log.Println(peer, womb.Name, txt)

			if isInList(txt, []string{"старт", "начать", "/старт", "/start", "start", "привет"}) {
				if isInUsers {
					sendMsg(fmt.Sprintf("Здравствуйте, %s!", womb.Name), peer, client)
				} else {
					sendMsg("Привет! Для того, чтобы ознакомиться с механикой бота, почитай справку https://vk.com/@wombat_bot-help (она также доступна по команде `помощь`. Чтобы завести вомбата, напиши `взять вомбата`. Приятной игры!", peer, client)
				}
			} else if isInList(txt, []string{"взять вомбата", "купить вомбата у арабов", "хочу вомбата"}) {
				if isInUsers {
					sendMsg("У тебя как бы уже есть вомбат лолкек. Если хочешь от него избавиться, то напиши `приготовить шашлык`", peer, client)
				} else {
					newWomb := User{
						Name:   standartNicknames[rand.Intn(len(standartNicknames))],
						XP:     0,
						Health: 5,
						Force:  2,
						Money:  10,
						Titles: map[uint16]Title{},
						Subs:   map[string]int64{},
					}
					_, err = users.Insert(&newWomb)
					checkerr(err)

					sendMsg(fmt.Sprintf("Поздравляю, у тебя появился вомбат! Ему выдалось имя `%s`. Ты можешь поменять имя командой `Поменять имя [имя]` за 3 монеты", womb.Name), peer, client)
				}
			} else if strings.HasPrefix(strings.ToLower(txt), "devtools") {
				if _, ok := womb.Titles[0]; ok {
					cmd := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(txt), "devtools"))
					if strings.HasPrefix(cmd, "set money") {
						strNewMoney := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(cmd), "set money"))
						if i, err := strconv.ParseUint(strNewMoney, 10, 64); err == nil {
							checkerr(err)
							womb.Money = i

							sendMsg(fmt.Sprintf("Операция проведена успешно! Шишей на счету: %d", womb.Money), peer, client)
						} else {
							sendMsg("Ошибка: неправильный синтаксис. Синтаксис команды: `devtools set money {кол-во шишей}`", peer, client)
						}
					} else if strings.HasPrefix(cmd, "reset") {
						arg := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(cmd), "reset"))
						switch arg {
						case "force":
							womb.Force = 2
							wombRes.Update(womb)
							sendMsg("Операция произведена успешно!", peer, client)
						case "health":
							womb.Health = 5
							wombRes.Update(womb)
							sendMsg("Операция произведена успешно!", peer, client)
						case "xp":
							womb.XP = 0
							wombRes.Update(womb)

							sendMsg("Операция произведена успешно!", peer, client)
						case "all":
							womb.Force = 2
							womb.Health = 5
							womb.XP = 0
							wombRes.Update(womb)

							sendMsg("Операция произведена успешно!", peer, client)
						default:
							sendMsg("Ошибка: неправильный синтаксис. Синтаксис команды: `devtools reset [force/health/xp/all]`", peer, client)
						}
					} else if cmd == "help" {
						sendMsg("https://vk.com/@wombat_bot-devtools", peer, client)
					}
				} else if strings.TrimSpace(txt) == "devtools on" {
					newTitle := Title{}
					titles.Find("{\"_id\":0}").One(&newTitle)
					womb.Titles[0] = newTitle
					wombRes.Update(womb)
					sendMsg("Выдан титул \"Вомботестер\" (ID: 0)", peer, client)
				}
			} else if isInList(txt, []string{"приготовить шашлык", "продать вомбата арабам", "слить вомбата в унитаз"}) {
				if isInUsers {
					if _, ok := womb.Titles[1]; !ok {
						err = wombRes.Delete()
						checkerr(err)
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
							wombRes.Update(womb)

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
						wombRes.Update(womb)

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
						wombRes.Update(womb)

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
						_, ok := womb.Titles[2]
						if ch := rand.Int(); ch%2 == 0 || ok && (ch%2 == 0 || ch%3 == 0) {
							rand.Seed(time.Now().UnixNano())
							win := rand.Intn(9) + 1
							womb.Money += uint64(win)
							sendMsg(fmt.Sprintf("Поздравляем! Вы нашли на дороге %d шишей! Теперь их у вас %d", win, womb.Money), peer, client)
						} else {
							sendMsg("Вы заплатили один шиш охранникам денежной дорожки, но увы, вы так ничего и не нашли", peer, client)
						}
						wombRes.Update(womb)

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
				} else if i, err := strconv.ParseInt(strID, 10, 64); err == nil {
					checkerr(err)
					ID := uint16(i)
					titleRes := titles.Find(fmt.Sprintf("{\"_id\":%d}", ID))
					rCount, err = titleRes.Count()
					checkerr(err)
					if rCount != 0 {
						elem := Title{}
						titleRes.One(&elem)
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
								rCount, err = users.Find(fmt.Sprintf("{\"_id\":%d}", ID)).Count()
								checkerr(err)
								if rCount != 0 {
									womb.Subs[args[1]] = ID
									wombRes.Update(womb)

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
					wombRes.Update(womb)

					sendMsg(fmt.Sprintf("Вы отписались от пользователя с алиасом %s", alias), peer, client)
				} else {
					sendMsg(fmt.Sprintf("Ошибка: вы не подписаны на пользователя с алиасом `%s`", alias), peer, client)
				}
			} else if isInList(txt, []string{"подписки", "мои подписки", "список подписок"}) {
				strSubs := "Вот список твоих подписок:"
				if len(womb.Subs) != 0 {
					for alias, id := range womb.Subs {
						res := users.Find(fmt.Sprintf("{\"_id\":%d}", id))
						rCount, err = res.Count()
						checkerr(err)
						if rCount != 0 {
							tWomb := User{}
							res.One(&tWomb)
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
					res := users.Find(fmt.Sprintf("{\"_id\":%d}", ID))
					rCount, err = res.Count()
					checkerr(err)
					if rCount != 0 {
						tWomb := User{}
						res.One(&tWomb)
						strTitles := ""
						for id, elem := range tWomb.Titles {
							strTitles += fmt.Sprintf("%s (ID: %d) | ", elem.Name, id)
						}
						strTitles = strings.TrimSuffix(strTitles, " | ")
						if strings.TrimSpace(strTitles) == "" {
							strTitles = "нет"
						}
						sendMsg(fmt.Sprintf("Вомбат  %s (ID: %d; Алиас: %s)\nТитулы: %s\n 🕳 %d XP \n ❤ %d здоровья \n ⚡ %d мощи \n 💰 %d шишей", tWomb.Name, ID, alias, strTitles, tWomb.XP, tWomb.Health, tWomb.Force, tWomb.Money), peer, client)
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
				} else if ID, err := strconv.ParseInt(strID, 10, 64); err == nil {
					res := users.Find(fmt.Sprintf("{\"_id\":%d}", ID))
					rCount, err = res.Count()
					checkerr(err)
					if rCount != 0 {
						tWomb := User{}
						res.One(&tWomb)
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
					res := users.Find(fmt.Sprintf("{\"_id\":%d}", womb.Subs[strID]))
					rCount, err = res.Count()
					checkerr(err)
					if rCount != 0 {
						tWomb := User{}
						res.One(&tWomb)
						strTitles := ""
						for id, elem := range tWomb.Titles {
							strTitles += fmt.Sprintf("%s (ID: %d) | ", elem.Name, id)
						}
						strTitles = strings.TrimSuffix(strTitles, " | ")
						if strings.TrimSpace(strTitles) == "" {
							strTitles = "нет"
						}
						sendMsg(fmt.Sprintf("Вомбат  %s (ID: %d; Алиас: %s)\nТитулы: %s\n 🕳 %d XP \n ❤ %d здоровья \n ⚡ %d мощи \n 💰 %d шишей", tWomb.Name, ID, strID, strTitles, tWomb.XP, tWomb.Health, tWomb.Force, tWomb.Money), peer, client)
					} else {
						sendMsg(fmt.Sprintf("Ошибка: неправильный алиас `%s` или не найден пользователь с ID %d. Обратитесь к @dikey_oficial, если такой вомбат есть", strID, womb.Subs[strID]), peer, client)
					}
				} else {
					sendMsg(fmt.Sprintf("Ошибка: не найден алиас `%s`", strID), peer, client)
				}
			} else if strings.HasPrefix(strings.ToLower(txt), "перевести шиши") {
				args := strings.Fields(strings.TrimSpace(strings.TrimPrefix(strings.ToLower(txt), "перевести шиши")))
				if len(args) < 2 {
					sendMsg("Ошибка: вы пропустили аргумент(ы). Синтаксис команды: `перевести шиши [кол-во] [ID/алиас получателя]`", peer, client)
				} else if len(args) > 2 {
					sendMsg("Ошибка: слишком много аргументов. Синтаксис команды: `перевести шиши [кол-во] [ID/алиас получателя]`", peer, client)
				} else {
					if amount, err := strconv.ParseUint(args[0], 10, 64); err == nil {
						if ID, err := strconv.ParseInt(args[1], 10, 64); err == nil {
							if womb.Money > amount {
								if amount != 0 {
									if ID != peer {
										res := users.Find(fmt.Sprintf("{\"_id\":%d}", ID))
										rCount, err = res.Count()
										checkerr(err)
										if rCount != 0 {
											tWomb := User{}
											res.One(&tWomb)
											womb.Money -= amount
											tWomb.Money += amount
											res.Update(tWomb)
											wombRes.Update(womb)

											sendMsg(fmt.Sprintf("Вы успешно перевели %d шишей на счёт %s. Теперь у вас %d шишей", amount, tWomb.Name, womb.Money), peer, client)
											sendMsg(fmt.Sprintf("Пользователь %s (ID: %d) перевёл вам %d шишей. Теперь у вас %d шишей", womb.Name, peer, amount, tWomb.Money), ID, client)
										} else {
											sendMsg(fmt.Sprintf("Ошибка: пользователя с ID %d не найдено", ID), peer, client)
										}
									} else {
										sendMsg("Ты читер блин нафиг!!!!!! нидам тебе самому себе перевести", peer, client)
									}
								} else {
									sendMsg("Ошибка: количество переводимых шишей должно быть больше нуля", peer, client)
								}
							} else {
								sendMsg(fmt.Sprintf("Ошибка: размер перевода (%d) должен быть меньше кол-ва ваших шишей (%d)", amount, womb.Money), peer, client)
							}
						} else if ID, ok := womb.Subs[args[1]]; ok {
							if womb.Money > amount {
								if amount != 0 {
									if ID == peer {
										sendMsg("Ты читер блин нафиг!!!!!! нидам тебе самому себе перевести", peer, client)
										continue
									}
									res := users.Find(fmt.Sprintf("{\"_id\":%d}", ID))
									rCount, err = res.Count()
									checkerr(err)
									if rCount != 0 {
										tWomb := User{}
										res.One(&tWomb)
										womb.Money -= amount
										tWomb.Money += amount
										res.Update(tWomb)
										wombRes.Update(womb)

										sendMsg(fmt.Sprintf("Вы успешно перевели %d шишей на счёт %s. Теперь у вас %d шишей", amount, tWomb.Name, womb.Money), peer, client)
										sendMsg(fmt.Sprintf("Пользователь %s (ID: %d) перевёл вам %d шишей. Теперь у вас %d шишей", womb.Name, peer, amount, tWomb.Money), ID, client)
									} else {
										sendMsg(fmt.Sprintf("Ошибка: пользователя с ID %d не найдено", ID), peer, client)
									}
								} else {
									sendMsg("Ошибка: количество переводимых шишей должно быть больше нуля", peer, client)
								}
							} else {
								sendMsg(fmt.Sprintf("Ошибка: размер перевода (%d) должен быть меньше кол-ва ваших шишей (%d)", amount, womb.Money), peer, client)
							}
						} else {
							sendMsg(fmt.Sprintf("Ошибка: алиас `%s` не обнаружено", args[1]), peer, client)
						}
					} else {
						if _, err := strconv.ParseInt(args[0], 10, 64); err == nil {
							sendMsg("Ошибка: количество переводимых шишей должно быть больше нуля", peer, client)
						} else {
							sendMsg("Ошибка: кол-во переводимых шишей быть числом", peer, client)
						}
					}
				}
			} else if txt == "обновить данные" && peer == 415610367 {
				loadUsers()
				sendMsg("Успешно обновлено!", peer, client)
			} else if isInList(txt, []string{"купить квес", "купить квесс", "купить qwess", "попить квес", "попить квесс", "попить qwess"}) {
				if isInUsers {
					if womb.Money >= 256 {
						if _, ok := womb.Titles[2]; !ok {
							titleRes := titles.Find("{\"_id\":2}")
							qwessTitle := Title{}
							titleRes.One(&qwessTitle)
							womb.Titles[2] = qwessTitle
							womb.Money -= 256
							wombRes.Update(womb)

							sendMsg("Вы купили чудесного вкуса квес у кролика-Лепса в ларьке за 256 шишей. Глотнув этот напиток, вы поняли, что получили новый титул с ID 2", peer, client)
						} else {
							womb.Money -= 256
							wombRes.Update(womb)

							sendMsg("Вы вновь купили вкусного квеса у того же кролика-Лепса в том же ларьке за 256 шишей. \"Он так освежает, я чувствую себя человеком\" — думаете вы. Ах, как вкусён квес!", peer, client)
						}
					} else {
						sendMsg("Вы подошли к ближайшему ларьку, но, увы, кролик-Лепс на кассе сказал, что надо 256 шишей, а у вас, к сожалению, меньше", peer, client)
					}
				} else {
					sendMsg("К сожалению, вам нужны шиши, чтобы купить квес, а шиши есть только у вомбатов...", peer, client)
				}
			}
		}
	}
}
