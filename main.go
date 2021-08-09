package main

import (
	env "github.com/joho/godotenv"
)

func main() {
	env.Load()
	bot, err := newBot()

	if err != nil {
		panic(err.Error())
	}

	bot.SetUpHandlers()
	bot.Start()
}
