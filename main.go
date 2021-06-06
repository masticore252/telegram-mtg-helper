package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	scryfall "github.com/BlueMonday/go-scryfall"
	env "github.com/joho/godotenv"
	tb "gopkg.in/tucnak/telebot.v2"
)

func main() {
	env.Load()
	bot, _ := makeBot()
	bot.Start()
}

func makeBot() (*tb.Bot, error) {

	// url that the web server will listen to this address
	port := os.Getenv("PORT")
	route := os.Getenv("ROUTE")
	listen := fmt.Sprintf(":%s/%s", port, route)

	// the webhook to be set to telegram API using setWebhook method
	webhook := os.Getenv("WEBHOOK_URL") + route

	Api := os.Getenv("TELEGRAM_API_URL")
	token := os.Getenv("TELEGRAM_TOKEN")
	isVerbose, _ := strconv.ParseBool(os.Getenv("VERBOSE_OUTPUT"))

	devMessage := "_\\(this bot is still in active development, reach to @masticore252 if you have any comments or suggestions\\)_"

	bot, _ := tb.NewBot(tb.Settings{
		URL:       Api,
		Token:     token,
		Verbose:   isVerbose,
		ParseMode: tb.ModeMarkdownV2,

		// Poller for getUpdates mode
		// Poller: &tb.LongPoller{Timeout: 10 * time.Second},

		// Poller for WebHook mode
		Poller: &tb.Webhook{
			Listen:   listen,
			Endpoint: &tb.WebhookEndpoint{PublicURL: webhook},
		},
	})

	// Handle inline queries
	bot.Handle(tb.OnQuery, func(q *tb.Query) {

		cards, _ := cardSearch(q.Text)

		results := make(tb.Results, len(cards))

		for i, card := range cards {

			if isDoubleFacedLayout(card.Layout) {

				for i, face := range card.CardFaces {
					singleResult := makeResultFromFace(face)
					results[i] = singleResult
					results[i].SetResultID(card.ID + fmt.Sprintf("-face-%v", i))
				}

			} else {

				result := makeResultFromCard(card)
				results[i] = result
				results[i].SetResultID(card.ID)

			}

		}

		err := bot.Answer(q, &tb.QueryResponse{
			Results:   results,
			CacheTime: 60, // a minute
		})

		if err != nil {
			log.Println(err)
		}
	})

	// Handle /start command
	bot.Handle("/start", func(m *tb.Message) {
		message := fmt.Sprint(
			devMessage+"\n\n",
			"Hi\\! I'm a Magic: the gathering bot\n\n",
			"I can help you find your favorite cards\n",
			"just open any of your chats and type\n\n",
			"_\"@MTGhelperBot Jace\"_ \\(or your favorite card name\\)\n\n",
			"I'll show a list of search results from scryfall\\.com\n",
			"tap one to preview it, then tap ✅ to send it\n",
			"Easy peasy\\!\n\n",
			"I support more complex querys, type /help to know more\n\n",
		)

		bot.Send(m.Chat, message)
	})

	// Handle /help command
	bot.Handle("/help", func(m *tb.Message) {
		bot.Send(m.Chat, devMessage)
	})

	// Handle all other messages
	bot.Handle(tb.OnText, func(m *tb.Message) {
		bot.Send(m.Chat, devMessage)
	})

	return bot, nil
}

func cardSearch(query string) ([]scryfall.Card, error) {
	client, err := scryfall.NewClient()

	if err != nil {
		return nil, err
	}

	context := context.Background()

	options := scryfall.SearchCardsOptions{
		// Unique:        scryfall.UniqueModeCards,
		// Order:         scryfall.OrderName,
		// Dir:           scryfall.DirAuto,
	}

	result, err := client.SearchCards(
		context,
		query,
		options,
	)

	if err != nil {
		return nil, err
	}

	// show no more than 50 results, as per telegram's limitation of inline queries
	max := 51

	if length := len(result.Cards); length < 50 {
		max = length
	}

	cards := result.Cards[0:max]

	return cards, nil
}

func isDoubleFacedLayout(layout scryfall.Layout) bool {
	doubleFacedLayouts := []scryfall.Layout{"modal_dfc", "transform", "double_faced_token", "art_series"}

	for _, val := range doubleFacedLayouts {
		if layout == val {
			return true
		}
	}

	return false
}

func makeResultFromCard(card scryfall.Card) *tb.PhotoResult {
	return &tb.PhotoResult{
		URL:      card.ImageURIs.Normal,
		ThumbURL: card.ImageURIs.Small,
	}
}

func makeResultFromFace(face scryfall.CardFace) *tb.PhotoResult {
	return &tb.PhotoResult{
		URL:      face.ImageURIs.Normal,
		ThumbURL: face.ImageURIs.Small,
	}
}
