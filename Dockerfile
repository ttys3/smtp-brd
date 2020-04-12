FROM golang:1.14.2-buster as builder

COPY . /build
WORKDIR /build

ARG DIST_MIRROR=no
ARG GOPROXY=direct
ARG LOCAL_PROXY

ENV TZ="Asia/Hong_Kong"

RUN set -eux; \
    	\
    echo "current architecture is: $(dpkg --print-architecture)"; \
    ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime; \
    echo "${TZ}" >  /etc/timezone; \
    echo "current date is: $(date)"; \
    http_proxy="${LOCAL_PROXY}" apt-get update && http_proxy="${LOCAL_PROXY}" apt-get install -y --no-install-recommends ca-certificates; \
	update-ca-certificates -f; \
    [ $DIST_MIRROR="yes" ] && sed -i 's|http://deb.debian.org|https://mirrors.ustc.edu.cn|g' /etc/apt/sources.list; \
    apt-get update; \
        \
    apt-get install -y --no-install-recommends make; \
    make release; \
    ./smtp-brd --version; \
    mkdir -p ./container/usr/local/bin; \
    cp -v ./smtp-brd ./container/usr/local/bin/smtp-brd; \
    cp -v ./config.toml ./container/etc/brd/

FROM 80x86/base-debian:buster-slim-amd64

ARG BUILD_DATE="n/a"
ARG VCS_REF="n/a"

LABEL org.label-schema.schema-version="1.0" \
org.label-schema.maintainer='HuangYeWuDeng <***@ttys3.net>' \
org.label-schema.name="80x86/smtp-brd" \
org.label-schema.description="smtp-brd" \
org.label-schema.build-date=$BUILD_DATE \
org.label-schema.vcs-ref=$VCS_REF \
org.label-schema.vcs-url="https://github.com/ttys3/smtp-brd" \
org.label-schema.url="https://ttys3.net" \
org.label-schema.vendor="ttys3"

COPY --from=builder /build/container/ /

ENV BRD_ADDR="0.0.0.0" \
    BRD_PORT="2525" \
    BRD_TLS=false \
    BRD_CERT="/etc/brd/ssl/ssl.crt" \
    BRD_KEY="/etc/brd/ssl/ssl.key" \
    BRD_DEBUG=false \
    BRD_USER="" \
    BRD_SECRET="" \
    BRD_PROVIDER="mailgun" \
    BRD_MAILGUN_API_KEY="" \
    BRD_MAILGUN_DOMAIN="" \
    BRD_MAILGUN_TIMEOUT=10 \
    BRD_SENDGRID_API_KEY="" \
    BRD_SENDGRID_TIMEOUT=10

WORKDIR /

EXPOSE 2525

CMD ["/usr/local/bin/smtp-brd"]
