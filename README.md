# zabbix-hook

```
export ZABBIX_USER=test
export ZABBIX_PASS=test
export ZABBIX_API=https://zabbix/zabbix/jsonrpc.php
export ZABBIX_CHATURL=https://mattermost/hooks/blablabla
export ZABBIX_HOSTGROUPS=test-channel
export ZABBIX_INTERVAL=60

go run main.go
```
