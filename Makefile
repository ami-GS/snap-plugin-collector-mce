default:
	$(MAKE) ondemand
ondemand:
	go build -o snap-plugin-collector-mce ./mce_ondemand/
stream:
	go build -o snap-plugin-collector-mce-stream ./mce_stream/
test:
	go test ./mce_ondemand/mce ./mce_stream/mce_stream
