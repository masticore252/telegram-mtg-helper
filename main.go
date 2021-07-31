package main

import (
	"os"
	"strconv"

	env "github.com/joho/godotenv"
	tb "gopkg.in/tucnak/telebot.v2"
)

func main() {
	env.Load()
	bot, _ := makeBot()
	bot.SetUpHandlers()
	bot.Start()
}

func makeBot() (*Bot, error) {

	bot := &Bot{}

	Api := os.Getenv("TELEGRAM_API_URL")
	token := os.Getenv("TELEGRAM_TOKEN")
	poller := bot.MakePoller(os.Getenv("POLLER_MODE"))
	isVerbose, _ := strconv.ParseBool(os.Getenv("VERBOSE_OUTPUT"))

	telebot, err := tb.NewBot(tb.Settings{
		URL:       Api,
		Token:     token,
		Poller:    poller,
		Verbose:   isVerbose,
		ParseMode: tb.ModeHTML,
	})

	if err != nil {
		return nil, err
	}

	bot.Bot = *telebot

	return bot, nil
}
