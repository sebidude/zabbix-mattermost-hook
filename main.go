package main

import (
	"crypto/tls"
	"fmt"
	"github.com/AlekSi/zabbix"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type filter map[string]interface{}

func main() {

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	httpclient := &http.Client{Transport: transport}
	chatClient := &http.Client{}

	user := os.Getenv("ZABBIX_USER")
	pass := os.Getenv("ZABBIX_PASS")
	apiUrl := os.Getenv("ZABBIX_API")
	chatUrl := os.Getenv("ZABBIX_CHATURL")
	hg_env := os.Getenv("ZABBIX_HOSTGROUPS")
	interval_str := os.Getenv("ZABBIX_INTERVAL")
	interval, _ := strconv.Atoi(interval_str)

	hg_list := strings.Split(hg_env, ",")
	api := zabbix.NewAPI(apiUrl)
	api.SetClient(httpclient)
	api.Login(user, pass)
	var hostgroup_names []string
	for _, v := range hg_list {
		hostgroup_names = append(hostgroup_names, v)
	}

	hostgroup_filter := filter{"name": hostgroup_names}
	hostgroups, err := api.Call("hostgroup.get", zabbix.Params{"output": "extend", "filter": hostgroup_filter})
	if err != nil {
		log.Println(err)
	}

	var hg_ids []string

	for _, hostgroup := range hostgroups.Result.([]interface{}) {
		groupid_str := hostgroup.(map[string]interface{})["groupid"].(string)
		hg_ids = append(hg_ids, groupid_str)
	}

	// remember the triggers with problems
	var problemTriggers []string
	var mustAdd bool

	problemTriggers = append(problemTriggers, "0")

	for {

		messages := []string{"---"}

		// get the triggers that have problems
		trigger_filter := filter{"value": "1", "status": "0"}
		triggers, _ := api.Call("trigger.get", zabbix.Params{
			"output":            "extend",
			"selectTags":        "extend",
			"groupids":          hg_ids,
			"expandDescription": "1",
			"filter":            trigger_filter})

		for _, trigger_raw := range triggers.Result.([]interface{}) {
			trigger := trigger_raw.(map[string]interface{})
			triggerid := trigger["triggerid"].(string)
			// check if we know the trigger
			for _, t := range problemTriggers {
				if t == triggerid {
					mustAdd = false
					break
				} else {
					mustAdd = true
				}
			}
			// we don't know the trigger, so we add it to the list.
			if mustAdd {
				problemTriggers = append(problemTriggers, triggerid)
				msg := fmt.Sprintf("[%s]: %s", "PROBLEM", trigger["description"])
				messages = append(messages, msg)
			}

		}

		// load all the known triggers with value 0 => resolved
		trigger_filter = filter{"value": "0", "status": "0"}
		triggers, _ = api.Call("trigger.get", zabbix.Params{
			"output":            "extend",
			"selectTags":        "extend",
			"expandDescription": "1",
			"triggerids":        problemTriggers,
			"filter":            trigger_filter})

		for _, trigger_raw := range triggers.Result.([]interface{}) {
			trigger := trigger_raw.(map[string]interface{})
			msg := fmt.Sprintf("[%s]: %s", "RESOLVED", trigger["description"])
			messages = append(messages, msg)
			for k, v := range problemTriggers {
				if trigger["triggerid"].(string) == v {
					problemTriggers = append(problemTriggers[:k], problemTriggers[k+1:]...)
				}
			}

		}

		// send all the messages via the webhook for the Incident channel

		if len(messages) < 2 {
			continue
		}

		msgPayload := ""
		for _, msg := range messages {
			log.Println(msg)
			msgPayload = fmt.Sprintf("%s%s\n", msgPayload, msg)
		}
		payload := fmt.Sprintf("{\"username\":\"Zabbix\",\"text\":\"%s\"}", msgPayload)
		form := url.Values{}
		form.Add("payload", payload)
		req, _ := http.NewRequest("POST", chatUrl, strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(form.Encode())))
		_, err := chatClient.Do(req)
		if err != nil {
			log.Printf("Cannot post to channel: %s", err)
		}

		time.Sleep(time.Duration(interval) * time.Second)

	}
}
