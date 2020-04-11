TARGET_EXEC_NAME = smtp-brd

DATE_VERSION := $(shell date +%Y%m%d)
GIT_VERSION := $(shell git rev-parse --short HEAD)

all: debug release

release: $(TARGET_EXEC_NAME)

debug: $(TARGET_EXEC_NAME).dbg

$(TARGET_EXEC_NAME):
	CGO_ENABLED=0 go build -o $@ -ldflags "-s -w -X main.Version=$(DATE_VERSION)@$(GIT_VERSION)" ./

$(TARGET_EXEC_NAME).dbg:
	go build -o $@ ./

buildah:
	sudo HTTP_PROXY=http://127.0.0.1:7070 buildah bud --network host --format=docker \
	--build-arg GOPROXY=https://goproxy.cn,direct \
	--build-arg HTTP_PROXY=http://127.0.0.1:7070 \
	--build-arg DIST_MIRROR=yes \
	--build-arg BUILD_DATE=$(DATE_VERSION) \
	--build-arg VCS_REF=$(GIT_VERSION) \
	-t 80x86/smtp-brd:latest ./

docker:
	sudo docker build --network host \
	--build-arg DIST_MIRROR=no \
	--build-arg BUILD_DATE=$(DATE_VERSION) \
	--build-arg VCS_REF=$(GIT_VERSION) \
	-t 80x86/smtp-brd:latest ./

clean:
	-rm -f $(TARGET_EXEC_NAME) $(TARGET_EXEC_NAME).dbg