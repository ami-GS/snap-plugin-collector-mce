ifdef TARGET_OS
  UNAME := $(TARGET_OS)
  OS := $(shell echo $(TARGET_OS) | tr A-Z a-z)
else
  UNAME := $(shell uname)
endif
build_dir := build/$(UNAME)/x86_64/

default:
	$(MAKE) ondemand
ondemand:
	GOOS=${OS} go build -o $(build_dir)snap-plugin-collector-mce ./mce_ondemand/
stream:
	GOOS=${OS} go build -o $(build_dir)snap-plugin-collector-mce-stream ./mce_stream/
test:
	go test ./mce_ondemand/mce ./mce_stream/mce_stream
