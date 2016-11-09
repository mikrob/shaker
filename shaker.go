package main

import (
	"fmt"
	"log"
	"net/http"
	"shaker/bots"
	"shaker/consul"

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
	var datas map[string][]bots.BotStatus

	for env, url := range envs {
		consul.NewClient(env)
		botList := consul.GetBotList(env)
		for _, bot := range botList {
			log.Println("BOT : ", bot)
			botStatusURL := fmt.Sprintf("%s/%s/status", url, bot)
			status := bots.RetrieveBotStatus(botStatusURL)
			wantedVersion := consul.GetBotVersionForEnv(env, bot)
			status.BotName = bot
			status.BotWantedVersion = wantedVersion
			if datas == nil {
				datas = make(map[string][]bots.BotStatus)
			}
			datas[env] = append(datas[env], status)
		}
	}

	fmt.Printf("%+v", datas)

	authorized.GET("/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title":  "BotsUnit Shaker",
			"status": datas,
		})
	})
	router.Run(":8080")
}
