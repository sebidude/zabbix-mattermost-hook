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

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("healthy"))
}

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
	problemIcon := os.Getenv("ZABBIX_PROBLEM_ICON")
	resolvedIcon := os.Getenv("ZABBIX_RESOLVED_ICON")
	interval, _ := strconv.Atoi(interval_str)

	hg_list := strings.Split(hg_env, ",")
	api := zabbix.NewAPI(apiUrl)
	api.SetClient(httpclient)
	_, err := api.Login(user, pass)
	if err != nil {
		log.Println(err)
	}
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
	log.Println(hostgroups)

	for _, hostgroup := range hostgroups.Result.([]interface{}) {
		groupid_str := hostgroup.(map[string]interface{})["groupid"].(string)
		hg_ids = append(hg_ids, groupid_str)
	}

	// remember the triggers with problems
	var problemTriggers []map[string]interface{}
	var mustAdd bool

	http.HandleFunc("/healthz", healthzHandler)
	go http.ListenAndServe(":8080", nil)

	for {

		messages := []string{}

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
			if len(problemTriggers) == 0 {
				mustAdd = true
			} else {

				for _, t := range problemTriggers {
					if t["triggerid"] == triggerid {
						mustAdd = false
						break
					} else {
						mustAdd = true
					}
				}
			}
			// we don't know the trigger, so we add it to the list.
			if mustAdd {
				problemTriggers = append(problemTriggers, trigger)
				msg := fmt.Sprintf("%s [%s]: %s", problemIcon, "PROBLEM", trigger["description"])
				messages = append(messages, msg)
			}

		}

		for i, rcount, rlen := 0, 0, len(problemTriggers); i < rlen; i++ {
			j := i - rcount
			mustDelete := true
			var trigger map[string]interface{}
			for _, trigger_raw := range triggers.Result.([]interface{}) {
				mustDelete = true
				trigger = trigger_raw.(map[string]interface{})
				if trigger["triggerid"].(string) == problemTriggers[j]["triggerid"] {
					mustDelete = false
					break
				}
			}
			if mustDelete {
				msg := fmt.Sprintf("%s [%s]: %s", resolvedIcon, "RESOLVED", problemTriggers[j]["description"])
				messages = append(messages, msg)
				problemTriggers = append(problemTriggers[:j], problemTriggers[j+1:]...)
				rcount++
			}
		}
		// send all the messages via the webhook for the Incident channel

		if len(messages) < 1 {
			time.Sleep(time.Duration(interval) * time.Second)
			continue
		}

		msgPayload := "---\n"
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
