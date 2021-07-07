package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
)

type pixivResponse struct {
	Pages []pixivPage `json:"body"`
}

type pixivPage struct {
	Urls struct {
		Original string `json:"original"`
	} `json:"urls"`
}

var pixivIDRegexp = regexp.MustCompile("[0-9]+")

func getOriginalImage(link string) (r *bytes.Reader, err error) {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return
	}
	req.Header.Set("Referer", "https://www.pixiv.net")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	r = bytes.NewReader(b)

	return
}

func getOriginalLinks(m string) (originals []string, err error) {
	link, err := getLinkFromID(m)
	if err != nil {
		return
	}

	resp, err := http.Get(link)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	page := &pixivResponse{}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(page)
	if err != nil {
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

func getLinkFromID(m string) (link string, err error) {
	id, err := getIDFromMessage(m)
	if err != nil {
		return
	}
	link = "https://www.pixiv.net/ajax/illust/" + id + "/pages"
	return
}

func getFilename(url string) string {
	return filepath.Base(url)
}
