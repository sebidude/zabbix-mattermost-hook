# zabbix-mattermost-hook
Yes, it is true, some people are using zabbix and mattermost.
Here is a dirty go app that loads firing triggers for a list of hostgroups and
sends out a notification via a mattermost webhook.

```
export ZABBIX_USER=test
export ZABBIX_PASS=test
export ZABBIX_API=https://zabbix/zabbix/jsonrpc.php
export ZABBIX_CHATURL=https://mattermost/hooks/blablabla
export ZABBIX_HOSTGROUPS="Public,Routers,Firewalls"
export ZABBIX_ICON_URL=https://here-goes-the-iconurl-for-the-bot.png
export ZABBIX_PROBLEM_ICON=":sos:"
export ZABBIX_RESOLVED_ICON=":white_check_mark:"

export ZABBIX_INTERVAL=60

go run main.go
```

Build a static binary
---------------------
```
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-d -s -w ' .
```
