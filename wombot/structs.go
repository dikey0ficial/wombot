package main

import (
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
	Money  uint32   `bson:"money"`
	Titles []uint16 `bson:"titles"`
	Sleep  bool     `bson:"sleep"`
}

// Attack реализует атаку
type Attack struct {
	ID   string `bson:"_id"`
	From int64  `bson:"from"`
	To   int64  `bson:"to"`
}

// Banked реализет вомбанковскую ячейку
type Banked struct {
	ID    int64  `bson:"_id"`
	Money uint32 `bson:"money"`
}

// ClanSettings реализует настройки клана
type ClanSettings struct {
	AviableToJoin     bool `bson:"aviable_to_join"`
	UseRequestsToJoin bool `bson:"use_requests_to_join"`
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

// Clan реализует клан
type Clan struct {
	Tag            string       `bson:"_id"`
	Name           string       `bson:"name"`
	Money          uint32       `bson:"money"` // Казна
	XP             uint32       `bson:"xp"`
	Leader         int64        `bson:"leader"`
	Banker         int64        `bson:"banker"`
	Members        []int64      `bson:"members"`
	Banned         []int64      `bson:"banned"`
	GroupID        int64        `bson:"group_id"`
	JoinRequests   []int64      `bson:"join_reqs"`
	LastRewardTime time.Time    `bson:"last_reward_time"`
	Settings       ClanSettings `bson:"settings"`
}

// SortedMembers возвращает отсортированный список участнииков
func (cl Clan) SortedMembers() []int64 {
	var membs = make([]int64, len(cl.Members))
	membs = cl.Members
	for i, id := range membs {
		if id == cl.Leader {
			membs[i], membs[0] = membs[0], cl.Leader
		} else if id == cl.Banker {
			if len(membs) > 1 {
				membs[i], membs[1] = membs[1], cl.Banker
			}
		}
	}
	return membs
}

// Laughter реализует структуру ржения
type Laughter struct {
	ChatID        int64     `bson:"_id"`
	Active        bool      `bson:"active"`
	LastStartTime time.Time `bson:"last_start_time"`
	Leader        int64     `bson:"leader"`
	Members       []int64   `bson:"members"`
}
