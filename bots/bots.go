package bots

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// BotStatus struct
type BotStatus struct {
	BotName           string
	BotRunningVersion string
	BotWantedVersion  string
}

//RetrieveBotStatus allow to retrieve a bot status with a given url
func RetrieveBotStatus(url string) BotStatus {
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
	var runningVersion string
	if bodyStr != "" {
		bodySplitted := strings.Split(bodyStr, "\n")
		if len(bodySplitted) > 1 {
			runningVersion = strings.Replace(bodySplitted[1], "version-", "", -1)
		} else {
			runningVersion = "Unknown"
		}
		//wantedVersion := strings.Replace(bodySplitted[2], "master_hash-", "", -1)
	}
	return BotStatus{
		BotRunningVersion: runningVersion,
	}
}
