package main

import (
	"fmt"
	"log"
	"net/http"
	"shaker/bots"
	"shaker/consul"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	envs = map[string]string{
		"integ-ufancyme": "https://integ-ufancyme.botsunit.io",
		"re7-ufancyme":   "https://re7-ufancyme.botsunit.io",
		"prod-ufancyme":  "https://beta.ufancyme.com"}
)

// simulate some private data
var secrets = gin.H{
	"admin": gin.H{"email": "infra@botsunit.com", "phone": "123433"},
}

func retrieveWantedVersion(env string, bot string, ch chan string) {
	wantedVersion := consul.GetBotVersionForEnv(env, bot)
	ch <- wantedVersion
}

func retrieveBotStatus(url string, ch chan bots.BotStatus) {
	status := bots.RetrieveBotStatus(url)
	ch <- status
}

func retrieveEnv(env string, url string, ch chan []bots.BotStatus) {
	consul.NewClient(env)
	botList := consul.GetBotList(env)
	result := make([]bots.BotStatus, len(botList))
	for index, bot := range botList {
		botStatusURL := fmt.Sprintf("%s/%s/status", url, bot)
		chStatus := make(chan bots.BotStatus)
		go retrieveBotStatus(botStatusURL, chStatus)
		ch := make(chan string)
		go retrieveWantedVersion(env, bot, ch)
		wantedVersion := <-ch
		status := <-chStatus
		status.BotName = bot
		status.BotWantedVersion = wantedVersion
		result[index] = status
	}
	fmt.Printf(" Result : %s", result)
	ch <- result
}

func getBotsDatas(c *gin.Context) {
	var datas map[string][]bots.BotStatus
	datas = make(map[string][]bots.BotStatus)
	//var wg sync.WaitGroup
	start := time.Now()
	//out := make(chan bots.BotStatus)
	//wg.Add(len(envs))
	for env, url := range envs {
		chStatusList := make(chan []bots.BotStatus)
		go retrieveEnv(env, url, chStatusList)
		envValues := <-chStatusList
		datas[env] = envValues
	}
	elapsed := time.Since(start)
	log.Printf("Retrieve data took %s", elapsed)
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title":  "BotsUnit Shaker",
		"status": datas,
	})
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

	authorized.GET("/index", getBotsDatas)
	router.Run(":8080")
}
