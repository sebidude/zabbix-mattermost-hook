# zabbix-hook

```
export ZABBIX_USER=test
export ZABBIX_PASS=test
export ZABBIX_API=https://zabbix/zabbix/jsonrpc.php
export ZABBIX_CHATURL=https://mattermost/hooks/blablabla
export ZABBIX_HOSTGROUPS="Public,Routers,Firewalls"
export ZABBIX_INTERVAL=60

go run main.go
```

Build a static binary
---------------------
```
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-d -s -w ' .
```
