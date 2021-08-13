package main

import (
	"log"
	"os"
	"path/filepath"
	"sync"
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
	_, err := b.Send(m.Chat, "Hello! "+
		"I will help you download original images from pixiv. "+
		"Just send me link.")
	if err != nil {
		log.Println(err)
		return
	}
}

func handleHelp(m *tb.Message) {
	_, err := b.Send(m.Chat, "The bot does not work with 18+ content. "+
		"The link should look like this: "+
		"pixiv.net/en/artworks/90539444, or just 90539444.",
		tb.NoPreview)
	if err != nil {
		log.Println(err)
		return
	}
}

func handleMessage(m *tb.Message) {
	id, err := getIDFromMessage(m.Text)
	if err != nil {
		log.Println(err)
		return
	}

	links, err := getOriginalLinks(id)
	if err != nil {
		_, err := b.Send(m.Chat, err.Error())
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(err)
		return
	}

	wg := sync.WaitGroup{}
	album := tb.Album{}

	for _, i := range links {
		wg.Add(1)

		go func(i string) {
			img, err := getOriginalImage(i)
			if err != nil {
				log.Println(err)
				return
			}

			f := &tb.Document{File: tb.FromReader(img),
				FileName: filepath.Base(i)}
			album = append(album, f)

			if len(album) == 10 {
				_, err = b.SendAlbum(m.Chat, album)
				if err != nil {
					log.Println(err)
					return
				}
				album = nil
			}

			wg.Done()
		}(i)

		wg.Wait()
	}

	if album != nil {
		_, err = b.SendAlbum(m.Chat, album)
		if err != nil {
			log.Println(err)
			return
		}
	}

}
