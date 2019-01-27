package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/labstack/echo"
	"golang.org/x/net/html"
)

type SetListItem struct {
	SetlistDataHTML string `json:"setlistdata"`
}

type ResponseData struct {
	Count int           `json:"count"`
	Data  []SetListItem `json:"data"`
}

type APIResponse struct {
	ErrorCode    int           `json:"error_code"`
	ErrorMessage string        `json:"error_message"`
	ResponseData *ResponseData `json:"response"`
}

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		names, err := getNames()
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("Error in GET %v", err))
		}
		return c.String(http.StatusOK, names)
	})

	e.Logger.Fatal(e.Start(":3434"))
}

func getNames() (string, error) {
	apikey := os.Getenv("PHISH_NET_API")

	url := fmt.Sprintf("https://api.phish.net/v3/setlist/random?apikey=%s", apikey)

	payload := strings.NewReader("{}")

	req, err := http.NewRequest("GET", url, payload)
	if err != nil {
		return "", err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	var apiResponse APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return "", err
	}

	doc, err := html.Parse(strings.NewReader(apiResponse.ResponseData.Data[0].SetlistDataHTML))
	if err != nil {
		return "", err
	}
	var f func(*html.Node)
	var songs []string
	pattern := regexp.MustCompile("http://phish.net/song/(.*)")
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				c := pattern.FindStringSubmatch(attr.Val)
				if c != nil {
					songs = append(songs, c[1])
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	randName := songs[rand.Intn(len(songs))] + "-" + songs[rand.Intn(len(songs))]
	return randName, nil
}