#!/bin/bash
################################################################

_fullscriptpath="$(readlink -f ${BASH_SOURCE[0]})"
BASEDIR="$(dirname $_fullscriptpath)"

# Parameters for the RPM
PREFIX="/"
MAINTAINER="sebidude"

if [ $# -gt 1 ]
then
    VERSION=$1
    RELEASE=$2
elif [ $# -eq 1 ]
then
    VERSION=$1
    RELEASE=1
fi

PKG_TYPE="rpm"
SRC_TYPE="dir"
PKG_NAME="zabbix-mattermost-hook"
VENDOR="github.com/sebidude"
DESCRIPTION="rpm for zabbix-mattermost-hook service"
URL="https://github.com/sebidude/zabbix-mattermost-hook"
PKG_ARCH="noarch"
EXCLUDE="*.svn*"

if [ -z "$GOPATH" ]
then
    export GOPATH=$HOME/go
fi

cd $BASEDIR
go get .
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-d -s -w ' .

# first do a cleanup
rm -rf build

# create the directories for the rpm structure
mkdir -p build/rpm/
mkdir -p build/rpm/usr/bin
mkdir -p build/rpm/etc/sysconfig
mkdir -p build/rpm/usr/lib/systemd/system

cp zabbix-mattermost-hook build/rpm/usr/bin
chmod 755 build/rpm/usr/bin/zabbix-mattermost-hook

cp zabbix-mattermost-hook.service build/rpm/usr/lib/systemd/system
cp zabbix-mattermost-hook.env build/rpm/etc/sysconfig

cat > postinstall.sh << EOF
#!/bin/bash
if ! id -u zabbix &>/dev/null
then
    useradd -g zabbix -s /bin/false zabbix
fi

chmod 400 /etc/sysconfig/zabbix-mattermost-hook.env
EOF

cd $BASEDIR/build/rpm

fpm \
    --prefix "$PREFIX" \
    -t "$PKG_TYPE" \
    -s "$SRC_TYPE" \
    -n "$PKG_NAME" \
    -a "$PKG_ARCH" \
    -m "$MAINTAINER" \
    --exclude "$EXCLUDE" \
    --description "$DESCRIPTION" \
    --vendor "$VENDOR" \
    --url "$URL" \
    --version "$VERSION" \
    --iteration "$RELEASE" \
    --after-install "$BASEDIR/postinstall.sh" \
    *

rm $BASEDIR/postinstall.sh

stat $BASEDIR/build/rpm/${PKG_NAME}-${VERSION}-${RELEASE}.${PKG_ARCH}.rpm
