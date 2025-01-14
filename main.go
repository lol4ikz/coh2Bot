package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/lol4ikz/coh2Bot/db"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const TG_APITOKEN string = "7615616860:AAFgMeWowwjgsPvWYfIRQwsoiLkgJOuspuo"

type subCommand int

const (
	rootCommand subCommand = iota
	teamsCommand
	matchCommand
)

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

// getting a random map
func getRandomMap() string {
	rand.Seed(time.Now().UnixNano())
	res := make(chan int)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Set the time as a seed value
		res <- (rand.Intn(13) + 1)
	}()
	go func() {
		wg.Wait()
		close(res)
	}()
	return coh2Maps[<-res]
}

func main() {
	newDb, err := db.NewDatabase("database.db")
	if err != nil {
		log.Printf("Cant create a new DB file")
	}
	defer newDb.Conn.Close()
	if err = newDb.CreateNewTables(); err != nil {
		log.Printf("Cant create tables")
	}
	bot, err := tgbotapi.NewBotAPI(TG_APITOKEN)
	if err != nil {
		panic(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	var subCmd subCommand = rootCommand

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	var lastRandomMap string

	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		if subCmd > 0 {
			switch subCmd {
			case teamsCommand:
				err = newDb.NewTeams(update.Message.Text)
				if err == nil {
					msg.Text = "✅ Команды добавлены"
				} else {
					msg.Text = "❌ Список команд полон или неправильный ввод"
				}
				subCmd = rootCommand
			case matchCommand:
				err = newDb.NewMatch(lastRandomMap, update.Message.Text)
				if err == nil {
					msg.Text = "✅ Счет добавлен"
				} else {
					msg.Text = "❌ Проверьте ввод"
				}
				subCmd = rootCommand
			}
		} else {
			switch update.Message.Command() {
			case "rand":
				lastRandomMap = getRandomMap()
				msg.Text = lastRandomMap
			case "teams":
				msg.Text = "Введите имена двух команд разделенные пробелами"
				subCmd = teamsCommand
			case "match":
				msg.Text = "Введите счет через тире \"2-0\""
				subCmd = matchCommand
			case "last":
				lastMatches, err := newDb.GetLastMatches()
				if err != nil {
					msg.Text = "Не удалось вывести статистику"
					log.Printf("%v", err)
				} else {
					log.Printf("%v", lastMatches)
					temp := fmt.Sprintf("%v", lastMatches)
					msg.Text = temp
				}

			default:
				continue
			}
		}
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}
