package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const TG_APITOKEN string = ""

var coh2Maps = map[int]string{
	1:  "Eindhoven Country",
	2:  "Rails and Metal",
	3:  "Road to Kharkov",
	4:  "Alliance of Defiance",
	5:  "Crossing in the Woods",
	6:  "Elst Outscirts",
	7:  "Fields of Winnekendonk",
	8:  "Minsk Pocket",
	9:  "Moscow Outscirts",
	10: "Belgorod",
	11: "Dreux Scout",
	12: "Highway to Baku",
	13: "Wofheze",
}

func getRandomMap() int {
	// Set the time as a seed value
	rand.Seed(time.Now().UnixNano())

	return rand.Intn(13) + 1

}

func main() {
	bot, err := tgbotapi.NewBotAPI(TG_APITOKEN)
	if err != nil {
		panic(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		if !update.Message.IsCommand() {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		switch update.Message.Command() {
		case "rand":
			msg.Text = coh2Maps[getRandomMap()]
		default:
			continue
		}
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
	fmt.Println("hello world ", coh2Maps[getRandomMap()])
}
