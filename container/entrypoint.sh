#!/bin/sh

# setup common things
. /opt/common.sh

CFG_FILE=/etc/brd/config.toml

# Default configuration file
if [ ! -f ${CFG_FILE} ]
then
	echo "init default config ..."
	cp /etc/default/config.toml ${CFG_FILE}
fi

exec /bin/s6-svscan -t0 /etc/services.d
