FROM scratch
COPY cacert.pem /etc/ssl/certs/ca-certificates.crt
ADD zabbix-mattermost-hook /
ENTRYPOINT ["/zabbix-mattermost-hook"]
