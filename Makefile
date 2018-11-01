build:
	go build -ldflags "-X main.buildTime=$(shell date '+%Y-%m-%dT%H:%M:%S')" -gcflags "-B"  rela_recommend
doc:
	godoc -http=:6060
