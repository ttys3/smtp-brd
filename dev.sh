#!/bin/sh

set -a
. ./.env.local
set +a

make clean && \
make debug && \
env BRD_USER=admin BRD_SECRET=admin \
./smtp-brd.dbg --debug