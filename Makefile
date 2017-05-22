all: treesql-server webui

start:
	go run server/server.go

start-dev-server:
	cd webui && PORT=9001 npm run start

deps:
	godep restore
	cd webui && npm install

webui:
	cd webui && npm run build

treesql-server:
	godep go build -o treesql-server server/server.go

clean:
	rm -r treesql-server
	rm -r webui/build

.PHONY: webui
