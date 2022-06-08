# Вомбот — Telegram-бот с вомбатами

Вомбот — простой Telegram-бот с лёгким RPG по типу [Жабабота](https://vk.com/toadbot)

Канал в Telegram: [@wombatobot_channel](https://t.me/wombatobot_channel); Сам бот: [@wombatobot](https://t.me/wombatobot)

## Запуск бота у себя

Для запуска бота Вам потребуется собрать его из исходного кода, запустить базу данных (MongoDB) и настроить бота

### Сборка из исходников

Необходимые программы:
  1. Go, версия 1.16 или выше
  2. Git (для клонирования репозитория, можно и без него)

Сборка:
 1. Клонируете репозиторий (`git clone https://github.com/dikey0ficial/wombot.git`)
 2. Заходите в папку `wombot` репозитория (`cd wombot/wombot`)
 3. Собираете с помощью `go build .`

### Конфигурация

Конфигурировать можно с помощью файла `config.toml` в папке с ботом либо переменных окружения. Ниже приведена таблица с описанием настроек

| Имя в toml      | Имя в переменных окружения | описание                                                                |
|:---------------:|:--------------------------:|------------------------------------------------------------------------:|
| tg_token        | TGTOKEN                    | Токен для бота                                                          |
| mongo_url       | MONGOURL                   | URI для подключения к MongoDB                                           |
| log_level       | LOGLEVEL                   | Уровень лога сообщений (0 — нет, 1 — только команды, 2 — все сообщения) |
| support_chat_id | SUPCHATID                  | ID чата поддержки (/admin, /support, /bug)                              |

### Настройка MongoDB для работы с ботом

Бесплатно получить базу данных можно получить на [atlas.mongodb.com](https://atlas.mongodb.com) (недоступно в РФ без VPN, блокируется со стороны MongoDB)

Единственное требование — наполнить коллекцию `imgs` наборами картинкок в следующем формате:

```json
{
  "_id":  "имя_набора",
  "imgs": [
    "id_картинки_1",
    "id_картинки_2"
  ]
}
```

Требуются следующие наборы:

| Имя набора         |                                  Где используется |
|:-------------------|--------------------------------------------------:|
| attacks            | Атака принять                                     |
| about              | О вомбате                                         |
| leps               | Купить квес (мало денег)                          |
| qwess              | Купить квес (достаточно денег; картинки с квесом) |
| sleep              | Спать                                             |
| unsleep            | Проснуться                                        |
| kill               | Приготовить шашлык                                |
| schweine           | Хрю                                               |
| new                | Завести вомбата                                   |
| cancel_0           | Атака отмена (у одной стороны)                    |
| cancel_1           | Атака отмена (у другой стороны)                   |

