package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

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

	Api := os.Getenv("TELEGRAM_API_URL")
	token := os.Getenv("TELEGRAM_TOKEN")
	poller := makePoller()
	isVerbose, _ := strconv.ParseBool(os.Getenv("VERBOSE_OUTPUT"))

	bot, _ := tb.NewBot(tb.Settings{
		URL:       Api,
		Token:     token,
		Poller:    poller,
		Verbose:   isVerbose,
		ParseMode: tb.ModeMarkdownV2,
	})

	devMessage := "_\\(this bot is still in active development, send a message to @masticore252 if you have any comments or suggestions\\)_"

	// Handle inline queries
	bot.Handle(tb.OnQuery, func(q *tb.Query) {

		cards, _ := cardSearch(q.Text)

		results := tb.Results{}

		for _, card := range cards {

			if isDoubleFacedLayout(card.Layout) {

				for index, face := range card.CardFaces {
					singleResult := newResultFromFace(face, card.ID, index)
					results = append(results, singleResult)
				}

			} else {

				singleResult := newResultFromCard(card)
				results = append(results, singleResult)

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
			"_\"@MTGhelperBot Jace\" \n\\(or your favorite card name\\)_\n\n",
			"I'll show a list of search results from scryfall\\.com, ",
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

func makePoller() tb.Poller {

	if pollerMode := os.Getenv("POLLER_MODE"); pollerMode == "webhook" {
		port := os.Getenv("PORT")
		route := os.Getenv("ROUTE")
		// url that the web server will listen to
		listen := fmt.Sprintf(":%s/%s", port, route)
		// the webhook to be set to telegram API using setWebhook method
		webhook := os.Getenv("WEBHOOK_URL") + route

		return &tb.Webhook{Listen: listen, Endpoint: &tb.WebhookEndpoint{PublicURL: webhook}}
	}

	return &tb.LongPoller{Timeout: 10 * time.Second}
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

func newResultFromCard(card scryfall.Card) *tb.PhotoResult {
	result := &tb.PhotoResult{
		URL:      card.ImageURIs.Normal,
		ThumbURL: card.ImageURIs.Small,
	}
	result.SetResultID(card.ID)
	return result
}

func newResultFromFace(face scryfall.CardFace, cardID string, faceIndex int) *tb.PhotoResult {
	result := &tb.PhotoResult{
		URL:      face.ImageURIs.Normal,
		ThumbURL: face.ImageURIs.Small,
	}
	faceID := fmt.Sprintf("%s-face-%d", cardID, faceIndex)
	result.SetResultID(faceID)
	return result
}
