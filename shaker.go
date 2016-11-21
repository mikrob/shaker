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
	ch <- result
}

type envCtx struct {
	BotsStatus      []bots.BotStatus
	SiteMetaVersion string
	EnvName         string
}

func getBotsDatas(c *gin.Context) {
	start := time.Now()
	//var datas map[string][]bots.BotStatus
	//datas = make(map[string][]bots.BotStatus)
	var ctx []envCtx
	// create a timeout chan that wait X sec and then send a timeout msg
	timeoutChan := make(chan bool, 1)
	go func() {
		time.Sleep(30 * time.Second)
		timeoutChan <- true
	}()
	for env, url := range envs {
		chStatusList := make(chan []bots.BotStatus)
		go retrieveEnv(env, url, chStatusList)
		var envValues []bots.BotStatus
		select {
		case envValuesCase := <-chStatusList:
			envValues = envValuesCase
			fmt.Println("received status")
		case <-timeoutChan:
			fmt.Println("Timeout for env : ", env)
		}
		//datas[env] = envValues
		envCtx := envCtx{EnvName: env}
		envCtx.BotsStatus = envValues
		envCtx.SiteMetaVersion = consul.GetSiteMetaVersion(env)
		ctx = append(ctx, envCtx)
	}
	elapsed := time.Since(start)
	log.Printf("Retrieve data took %s", elapsed)
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title": "BotsUnit Shaker",
		//"status": datas,
		"time": elapsed,
		"ctx":  &ctx,
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

	authorized.GET("/", getBotsDatas)
	router.Run(":8080")
}
