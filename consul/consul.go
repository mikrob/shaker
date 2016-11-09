package consul

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/consul/api"
)

var (
	consulClient *api.Client
)

//NewClient allow to initialize a new consul client
func NewClient(env string) {
	var consulURL string
	if os.Getenv("CONSUL") != "" {
		consulURL = os.Getenv("CONSUL")
	} else {
		consulURL = "localhost:8500"
	}

	config := api.Config{
		Datacenter: EnvToDc(env),
		Scheme:     "http",
		Address:    consulURL,
	}
	var err error
	consulClient, err = api.NewClient(&config)
	if err != nil {
		fmt.Println(fmt.Errorf("Error while initializing consul client : %s", err.Error()))
	}
}

//GetBotVersionForEnv Return bot version for a given bot in a given env
func GetBotVersionForEnv(env string, bot string) string {
	// Get a handle to the KV API
	kv := consulClient.KV()
	// Lookup the pair
	keyPath := fmt.Sprintf("%s/bots_versions/%s/version", env, bot)
	pair, _, errKV := kv.Get(keyPath, nil)
	if errKV != nil {
		fmt.Println(fmt.Errorf("[ERROR] Error while calling consul API : %s", errKV.Error()))
		return ""
	}
	return string(pair.Value)
}

//GetBotList list existing bot in bot_version folder in kv store
func GetBotList(env string) []string {
	var botList []string
	kv := consulClient.KV()
	prefix := fmt.Sprintf("%s/bots_versions/", env)
	pairs, _, err := kv.Keys(prefix, "/", nil)
	if err != nil {
		fmt.Println(fmt.Errorf("[ERROR] Error while calling consul API : %s", err.Error()))
		return []string{}
	}
	for _, pair := range pairs {
		botSplitted := strings.Split(pair, "/")
		bot := botSplitted[2]
		botList = append(botList, bot)
	}
	return botList
}

//EnvToDc convert an env name to a dc name
func EnvToDc(env string) string {
	envSplitted := strings.Split(env, "-")
	dcName := fmt.Sprintf("dc1%s%s", envSplitted[0], envSplitted[1])
	return dcName
}
