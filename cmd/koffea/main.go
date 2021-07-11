package main

import (
	"fmt"
	"log"

	"github.com/ikaruswill/koffea/client/sqlite"
	"github.com/ikaruswill/koffea/client/telegram"
	"github.com/ikaruswill/koffea/config"
	"github.com/ikaruswill/koffea/storage"
	tb "gopkg.in/tucnak/telebot.v2"
)

func main() {
	koffeaConfig := &config.Config{}

	err := koffeaConfig.Load("config.yaml")

	if err != nil {
		log.Fatal(err)
		return
	}

	tClient, err := telegram.NewConnection(koffeaConfig.Telegram)

	if err != nil {
		log.Fatal(err)
		return
	}

	// var dbClient interface{}

	// switch koffeaConfig.Storage.Driver {
	// case config.SqliteStorageDriver:
	// 	dbClient, err = sqlite.NewConnection(koffeaConfig.Storage.Sqlite)
	// case config.PostgresStorageDriver:
	// 	dbClient, err = sqlite.NewConnection(koffeaConfig.Storage.Postgres)
	// }

	var dbClient *sqlite.Client
	dbClient, err = sqlite.NewConnection(koffeaConfig.Storage.Sqlite)

	dbClient.DB.AutoMigrate(
		&storage.User{},
		&storage.Order{},
		&storage.GroupBuy{},
	)

	if err != nil {
		log.Fatal(err)
		return
	}

	tClient.Bot.Handle("/start", func(m *tb.Message) {
		user := &storage.User{
			FirstName:    m.Sender.FirstName,
			LastName:     m.Sender.LastName,
			Username:     m.Sender.Username,
			LanguageCode: m.Sender.LanguageCode,
			IsBot:        m.Sender.IsBot,
		}
		dbClient.DB.FirstOrCreate(&user)

		tClient.Bot.Send(m.Sender, fmt.Sprintf("Hi %s! My name is Koffea, how can I help you today?", user.FirstName))
		tClient.Bot.Send(m.Sender, `/listgb - List active group buys
		/viewgb - View information about a group buy
		/startgb - Start a group buy
		/joingb - Join a group buy
		/endgb - End a group buy`)
	})

	tClient.Bot.Handle("/listgb", func(m *tb.Message) {
		var groupBuys []storage.GroupBuy
		dbClient.DB.Order("created_at desc").Find(&groupBuys)

		response := "Here are the current active group buys:\n"
		for idx, groupBuy := range groupBuys {
			user := &storage.User{}
			dbClient.DB.Where("id = ?", groupBuy.ID).First(&user)
			response += fmt.Sprintf("%d. [%s] %s by %s %s", idx+1, groupBuy.CreatedAt.Format("2006-01-02"), "SHOP NAME", user.FirstName, user.LastName)
		}
		tClient.Bot.Send(m.Sender, response)
	})

	tClient.Bot.Start()
}
