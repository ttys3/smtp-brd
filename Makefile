TARGET_EXEC_NAME = smtp-brd

all: debug release

release: $(TARGET_EXEC_NAME)

debug: $(TARGET_EXEC_NAME).dbg

$(TARGET_EXEC_NAME):
	go build -o $@ -ldflags "-s -w -X main.Version=1.0.0" ./

$(TARGET_EXEC_NAME).dbg:
	go build -o $@ ./

clean:
	-rm -f $(TARGET_EXEC_NAME) $(TARGET_EXEC_NAME).dbg