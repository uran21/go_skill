package main

import (
	"log"
	"strings"
	"fmt"
	"strconv"
	"net/http"
	"encoding/json"
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)
type binanceResp struct {
	Price float64 `json:"price,string"`
	Code  int64 	`json:"code"`
}

func get_key() string{
return	"2130453542:AAEHpTguBThnLr4hjYLQ1Q6t7zUOAnoklW4"
}

type wallet map[string]float64
var db = map[int64]wallet{}

func main() {
	bot, err := tgbotapi.NewBotAPI(get_key())
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Println(update.Message.Text)

		msgArr:=strings.Split(update.Message.Text, " ")

		switch msgArr[0] {
		case "ADD":
			if msgArr[0]==""&&msgArr[1]=="" {continue} //не смог исправить вылета программы при не дописовании аргументов ЖДУ ПОДСКАЗСКИ
			summ, err := strconv.ParseFloat(msgArr[2], 64)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "impossible conversion summ"))
				continue
			}
			if _, ok := db[update.Message.Chat.ID]; !ok {
				db[update.Message.Chat.ID]=wallet{}
			}
			db[update.Message.Chat.ID][msgArr[1]] += summ
			msg:=fmt.Sprintf("Balance: %s %f", msgArr[1], db[update.Message.Chat.ID][msgArr[1]])
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
		case "SUB":
			summ, err := strconv.ParseFloat(msgArr[2], 64)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "impossible conversion summ"))
				continue
			}
			if _, ok := db[update.Message.Chat.ID]; !ok {
				db[update.Message.Chat.ID]=wallet{}
			}
			db[update.Message.Chat.ID][msgArr[1]] -= summ
			msg:=fmt.Sprintf("Balance: %s %f", msgArr[1], db[update.Message.Chat.ID][msgArr[1]])
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
		case "DEL":
			delete(db[update.Message.Chat.ID], msgArr[1])
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Curency Deleted"))
		case "SHOW":
			msg:="Balance:\n"
			var usdSumm float64
			var rubSumm float64
			//здесь втыкаемся в код и конвертируем сумму
			rubSumm =getRub()
			for key, value :=range db[update.Message.Chat.ID] {
			coinPrice, err := getPrice(key)
			if err != nil{
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
			}
			if value<=0{ msg +=	fmt.Sprintf("Can not be negative\n ") continue} //ЕСЛИ МИНУС
			usdSumm += value * coinPrice
			msg +=	fmt.Sprintf("%s: %f [%.2f $]  [%.2f rub]\n", key, value, value*coinPrice, (value*coinPrice)*rubSumm)
			}
			msg +=	fmt.Sprintf("Summ: %.2f $ and in Rub %.2f\n",usdSumm, usdSumm*rubSumm)

			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
		default:
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Not known command"))
}



	}
}
//Здесь получаем рубль
func getRub() (price float64){
	resp, err := http.Get("https://api.binance.com/api/v3/ticker/price?symbol=USDTRUB")
	if err !=nil {
		return
	}
	defer resp.Body.Close()

	var jsonResp binanceResp
	err = json.NewDecoder(resp.Body).Decode(&jsonResp)
	if err !=nil {
		return
	}
	if jsonResp.Code !=0{
		err = errors.New("Not correct curence")
		return
	}
	price = jsonResp.Price
	log.Println(price)
	return
}

func getPrice(coin string) (price float64, err error){
resp, err := http.Get(fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%sUSDT", coin))
if err !=nil {
	return
}
defer resp.Body.Close()

var jsonResp binanceResp
err = json.NewDecoder(resp.Body).Decode(&jsonResp)
if err !=nil {
	return
}
if jsonResp.Code !=0{
	err = errors.New("Not correct curence")
	return
}
price = jsonResp.Price
return
}
