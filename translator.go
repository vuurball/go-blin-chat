package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type Sentence struct {
	Trans string `json:"trans"`
	//Orig	    string `json:"orig"`
	//Backend   int    `json:"backend"`
}
type apiResponse struct {
	Sentences []Sentence `json:"sentences"`
	//Src      string      `json:"src"`
}

func TranslateMessage(msg Message) string {

	apiurl := "https://translate.googleapis.com/translate_a/single?client=gtx&ie=UTF-8&oe=UTF-8&sl=" + msg.SourceLang + "&tl=" + msg.TargetLang + "&hl=ru&dt=t&dj=1&source=icon&tk=467103.467103&q="

	escapedMessage := url.QueryEscape(msg.Message)

	resp, err := http.Get(apiurl + escapedMessage)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var r apiResponse
	err1 := json.Unmarshal(bytes, &r)
	if err1 != nil {
		log.Printf("err was %v", err1)
	}

	return r.Sentences[0].Trans
}
