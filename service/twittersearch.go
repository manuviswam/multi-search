package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	m "github.com/manuviswam/multisearch/model"
)

const (
	twitterOAuthUrl = "https://api.twitter.com/oauth2/token"
	twitterUrl      = "https://api.twitter.com/1.1/search/tweets.json?count=1&q=%s"
)

type TwitterSearch struct {
	bearerToken string
}

func (t *TwitterSearch) SetBearerToken(encodedTwitterKey string) {
	client := &http.Client{}
	requestBody := bytes.NewReader([]byte("grant_type=client_credentials"))
	req, err := http.NewRequest("POST", twitterOAuthUrl, requestBody)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", encodedTwitterKey))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	twitterToken := m.TwitterTokenResponse{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&twitterToken)
	if err != nil {
		panic(err)
	}

	t.bearerToken = twitterToken.AccessToken
}

func (t *TwitterSearch) Search(query string, c chan m.SearchResult) {
	defer close(c)
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf(twitterUrl, query), nil)
	if err != nil {
		c <- m.SearchResult{
			Error: err.Error(),
		}
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", t.bearerToken))
	start := time.Now()
	resp, err := client.Do(req)
	fmt.Println("Elapsed time for twitter ", time.Since(start))
	defer resp.Body.Close()
	if err != nil {
		c <- m.SearchResult{
			Error: err.Error(),
		}
		return
	}

	twitterResponse := m.TwitterResponse{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&twitterResponse)
	if err != nil {
		c <- m.SearchResult{
			Error: err.Error(),
		}
		return
	}

	if len(twitterResponse.Statuses) < 1 {
		c <- m.SearchResult{
			Error: "No response obtained",
		}
		return
	}
	c <- m.SearchResult{
		Text: twitterResponse.Statuses[0].Text,
	}
}
