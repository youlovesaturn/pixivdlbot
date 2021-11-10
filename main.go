package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

var b *tb.Bot
var err error
var ctx, cancel = context.WithTimeout(context.Background(), time.Minute*10)

func main() {
	b, err = tb.NewBot(tb.Settings{
		URL:    os.Getenv("API_URL"),
		Token:  os.Getenv("TOKEN"),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
		return
	}

	defer cancel()

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
		_, _ = b.Send(m.Chat, err.Error())
		log.Println(err)
		return
	}

	wg := sync.WaitGroup{}
	results := make(chan result)
	wg.Add(len(links))
	go func() { wg.Wait(); close(results) }()

	fetch := func(link string) {
		defer wg.Done()
		img, errImg := fetchImage(ctx, link)
		if errImg != nil {
			err := fmt.Errorf("%q: %w", link, errImg)
			results <- result{err: err}
			return
		}
		results <- result{img: img}
	}

	for _, link := range links {
		go fetch(link)
	}
	collectAlbums(ctx, results, m)
}

type image struct {
	Data     []byte
	Filename string
}

type result struct {
	img image
	err error
}

func fetchImage(ctx context.Context ,link string) (img image, err error) {
	img.Data, err = getOriginalImage(link)
	img.Filename = filepath.Base(link)
	if err != nil {
		log.Println(err)
		return
	}
	return
}

func collectAlbums(ctx context.Context, results <-chan result, m *tb.Message) {
	album := tb.Album{}

	send := func() {
		_, err := b.SendAlbum(m.Chat, album)
		if err != nil {
			log.Println(err)
			return
		}
	}

	for r := range results {
		switch {
		case r.err != nil:
			log.Printf("fetching image: %v", r.err)
		default:
			f := &tb.Document{File: tb.FromReader(bytes.NewReader(r.img.Data)),
				FileName: r.img.Filename}
			album = append(album, f)
			if len(album) == 10 {
				send()
				album = nil
			}
		}
	}
	if album != nil {
		send()
	}
}