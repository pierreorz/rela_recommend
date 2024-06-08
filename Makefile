build:
	# cat requirements.txt | sed '/^$//d' | go get -v
	go build -ldflags "-X main.buildTime=$(shell date '+%Y-%m-%dT%H:%M:%S')" -gcflags "-B"  rela_recommend
doc:
	godoc -http=:6060
generate-doc:
	GO111MODULE=off go get -u github.com/go-swagger/go-swagger/cmd/swagger && swagger generate spec -o ./static/swagger.json --scan-models
generate-doc-builtin:
	GO111MODULE=off swagger generate spec -o ./static/swagger.json --scan-models
