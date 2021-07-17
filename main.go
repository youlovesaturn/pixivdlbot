package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

var b *tb.Bot

func main() {
	var err error

	b, err = tb.NewBot(tb.Settings{
		URL:    os.Getenv("API_URL"),
		Token:  os.Getenv("TOKEN"),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle("/start", handleStart)
	b.Handle("/help", handleHelp)
	b.Handle(tb.OnText, handleMessage)
	b.Start()

}

func handleStart(m *tb.Message) {
	b.Send(m.Chat, "Hello! "+
		"I will help you download original images from pixiv. "+
		"Just send me link.")
}

func handleHelp(m *tb.Message) {
	b.Send(m.Chat, "The bot does not work with 18+ content. "+
		"The link should look like this: "+
		"pixiv.net/en/artworks/90539444, or just 90539444.",
		tb.NoPreview)
}

func handleMessage(m *tb.Message) {
	id, err := getIDFromMessage(m.Text)
	if err != nil {
		b.Send(m.Chat, err.Error())
		return
	}

	links, err := getOriginalLinks(id)
	if err != nil {
		b.Send(m.Chat, err.Error())
		return
	}

	for _, i := range links {

		go func(i string) {

			img, err := getOriginalImage(i)
			if err != nil {
				return
			}

			f := &tb.Document{File: tb.FromReader(img),
				FileName: filepath.Base(i)}
			b.Send(m.Chat, f)

		}(i)
	}

}
