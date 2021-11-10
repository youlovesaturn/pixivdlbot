package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"time"
)

type pixivResponse struct {
	Error        bool        `json:"error"`
	ErrorMessage string      `json:"message"`
	Pages        []pixivPage `json:"body"`
}

type pixivPage struct {
	Urls struct {
		Original string `json:"original"`
	} `json:"urls"`
}

var errorTranslations = map[string]string{
	"該当作品は削除されたか、存在しない作品IDです。": "The work ID has been deleted or does not exist.",
}

var pixivIDRegexp = regexp.MustCompile("[0-9]+")

func translateError(japanese string) string {
	english, ok := errorTranslations[japanese]
	if !ok {
		return japanese
	}
	return english
}

func getOriginalImage(link string) (b []byte, err error) {
	client := &http.Client{
		Timeout: time.Minute * 5,
	}

	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Set("Referer", "https://www.pixiv.net")

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Println(err)
			return
		}
	}()

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}
	return
}

func getOriginalLinks(id string) (originals []string, err error) {
	link := "https://www.pixiv.net/ajax/illust/" + id + "/pages"

	resp, err := http.Get(link)
	if err != nil {
		log.Println(err)
		return
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Println(err)
			return
		}
	}()

	page := &pixivResponse{}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(page)
	if err != nil {
		log.Println(err)
		return
	}

	if page.Error {
		err = errors.New(translateError(page.ErrorMessage))
		return
	}

	for _, l := range page.Pages {
		originals = append(originals, l.Urls.Original)
	}

	return
}

func getIDFromMessage(m string) (id string, err error) {
	id = pixivIDRegexp.FindString(m)
	if id == "" {
		return "", errors.New("no result")
	}
	return

}
