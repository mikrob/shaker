package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// simulate some private data
var secrets = gin.H{
	"admin":  gin.H{"email": "infra@botsunit.com", "phone": "123433"},
	"mikrob": gin.H{"email": "mikael.robert@botsunit.com", "phone": "666"},
	"benj":   gin.H{"email": "benjamin.jorand@botsunit.com", "phone": "523443"},
}

// BotStatus struct
type BotStatus struct {
	BotRunningVersion string
	BotWantedVersion  string
}

func retrieveBotStatus(url string) BotStatus {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return BotStatus{}
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		return BotStatus{}
	}

	defer resp.Body.Close()

	bs, errRead := ioutil.ReadAll(resp.Body)
	if errRead != nil {
		log.Fatal("Error while reading response body :", err.Error())
	}
	bodyStr := string(bs)
	bodySplitted := strings.Split(bodyStr, "\n")
	runningVersion := strings.Replace(bodySplitted[1], "version-0.0.1-", "", -1)
	//wantedVersion := strings.Replace(bodySplitted[2], "master_hash-", "", -1)
	return BotStatus{
		BotRunningVersion: runningVersion,
		BotWantedVersion:  "42", //wantedVersion,
	}
}

func main() {
	router := gin.Default()
	router.Static("/css", "./css")
	router.Static("/js", "./js")
	router.Static("/fonts", "./fonts")
	authorized := router.Group("/", gin.BasicAuth(gin.Accounts{
		"admin":  "password",
		"mikrob": "password",
		"benj":   "password",
	}))

	router.LoadHTMLGlob("templates/*")

	firstBotURL := "https://integ-ufancyme.botsunit.io/status"
	status := retrieveBotStatus(firstBotURL)

	authorized.GET("/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title":  "BotsUnit Shaker",
			"status": status,
		})
	})
	router.Run(":8080")
}
