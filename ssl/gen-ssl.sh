#!/bin/sh

openssl req -new -newkey rsa:2048 \
-sha256 \
-days 3650 \
-nodes \
-x509 \
-out ssl.crt \
-keyout ssl.key \
-subj "/C=US/ST=CA/L=LA/O=ttys3/OU=smtp-brd for remark42/CN=remark42.smtpd-brd.localhost"
