package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/lol4ikz/coh2Bot/db"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const TG_APITOKEN string = "7615616860:AAFgMeWowwjgsPvWYfIRQwsoiLkgJOuspuo"

type BotWrapper struct {
	bot *tgbotapi.BotAPI
}

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

// generating a PDF file with last 10 match results
func getPdf(dB *db.Database) (*bytes.Buffer, error) {
	matches, err := dB.GetLastMatches()
	if err != nil {
		return nil, err
	}
	var res bytes.Buffer
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 15)

	// Add headers for the columns
	pdf.CellFormat(10, 10, "ID", "1", 0, "C", false, 0, "")
	pdf.CellFormat(50, 10, "Map", "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 10, "Team 1", "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 10, "Team 2", "1", 0, "C", false, 0, "")
	pdf.CellFormat(20, 10, "Score 1", "1", 0, "C", false, 0, "")
	pdf.CellFormat(20, 10, "Score 2", "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 10, "Date", "1", 1, "C", false, 0, "")

	// Set font for the match results
	pdf.SetFont("Arial", "", 12)
	for _, match := range matches {
		pdf.CellFormat(10, 10, strconv.Itoa(match.ID), "1", 0, "C", false, 0, "")
		pdf.CellFormat(50, 10, match.MAP, "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 10, match.TEAM1, "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 10, match.TEAM2, "1", 0, "C", false, 0, "")
		pdf.CellFormat(20, 10, strconv.Itoa(match.SCORE1), "1", 0, "C", false, 0, "")
		pdf.CellFormat(20, 10, strconv.Itoa(match.SCORE2), "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 10, match.DATE, "1", 1, "C", false, 0, "")
	}
	err = pdf.Output(&res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// universal sender for text messages and doc messages
func (b *BotWrapper) sender(m *tgbotapi.MessageConfig, d *tgbotapi.DocumentConfig) error {
	var err error
	if m != nil && m.Text != "" {
		_, err = b.bot.Send(m)
	} else if d != nil {
		_, err = b.bot.Send(d)
	} else {
		return fmt.Errorf("both message and document are nil or empty")
	}
	return err
}

func main() {
	newDb, err := db.NewDatabase("database.db")
	if err != nil {
		log.Panic("Cant create a new DB file")
	}
	defer newDb.Close()
	bot, err := tgbotapi.NewBotAPI(TG_APITOKEN)
	if err != nil {
		panic(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	userID := int64(1636459796)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	bw := &BotWrapper{bot: bot}
	var subCmd subCommand = rootCommand
	var lastRandomMap string
	// seed value for math/rand
	rand.Seed(time.Now().UnixNano())

	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		doc := tgbotapi.NewDocument(update.Message.Chat.ID, nil)
		if subCmd != rootCommand {
			switch subCmd {
			case teamsCommand:
				err = newDb.NewTeams(update.Message.Text)
				if err == nil {
					msg.Text = "✅ Команды добавлены"
				} else {
					log.Panic(err)
					msg.Text = "❌ Список команд полон или неправильный ввод"
				}
				subCmd = rootCommand
			case matchCommand:
				err = newDb.NewMatch(lastRandomMap, update.Message.Text)
				if err == nil {
					msg.Text = "✅ Счет добавлен"
				} else {
					log.Panic("%v\n", err)
					msg.Text = "❌ Проверьте ввод"
				}
				subCmd = rootCommand
			}
		} else {
			if !update.Message.IsCommand() {
				continue
			}
			switch update.Message.Command() {
			case "new":
				if update.Message.From.ID != userID {
					msg.Text = "❌ Только админ может начать новый турнир"
					break
				}
				if err = newDb.CreateNewTables(); err != nil {
					log.Panicf("Cant create tables: %v", err)
					msg.Text = "❌ Не удалось создать турнир"
				} else {
					msg.Text = "✅ Все результаты обнулены начат новый турнир"
				}
			case "rand":
				lastRandomMap = getRandomMap()
				msg.Text = lastRandomMap
			case "teams":
				msg.Text = "Введите имена двух команд разделенные пробелами"
				subCmd = teamsCommand
			case "match":
				if newDb.TeamsTableIsEmpty() || lastRandomMap == "" {
					msg.Text = "❌ Добавьте команды и/или выберите карты"
					break
				}
				msg.Text = "Введите счет через тире \"2-0\""
				subCmd = matchCommand
			case "sum":
				total1, total2, err := newDb.GetTotalScores()
				if err != nil {
					log.Panicf("%v\n", err)
					msg.Text = "Нет результатов, или произошла ошибка"
				} else {
					msg.Text = fmt.Sprintf("Первая команда: %d\nВторая команда: %d", total1, total2)
				}
			case "last":
				pdfBuffer, err := getPdf(newDb)
				if err != nil {
					log.Panicf("%v\n", err)
					continue
				}
				// preparing a file to the upload
				file := tgbotapi.FileBytes{
					Name:  "lastMatches.pdf",
					Bytes: pdfBuffer.Bytes(),
				}
				doc.BaseFile.File = file
				doc.Caption = "Результаты последних 10 игр"
				if _, err := bot.Send(doc); err != nil {
					log.Panic(err)
				}
				continue
			default:
				msg.Text = "===Доступные команды==\n" +
					"/new 	- Удалить все результаты и начать заново\n"+                   
					"/rand 	- Выбрать карту\n"+             
					"/teams - Добавить команды\n"+                
					"/match - Добавить статистику матча\n"+                        
					"/last 	- Счет последних игр\n"+                 
					"/sum 	- Общий счет команд"
			}
		}
		if err := bw.sender(&msg, &doc); err != nil {
			log.Panic(err)
		}
	}
}
