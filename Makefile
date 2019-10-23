#
# Makefile for ayame, WebRTC signaling server
#
PROG=ayame
VERSION=19.04.05
# -----------------------------------------------------------------------------------------------------------------------
usage:
	@echo "WebRTC signaling server : $(PROG) $(VERSION)"
	@echo "> make [build|run|kill|docker|compose|ngrok|git]"
# -----------------------------------------------------------------------------------------------------------------------
build b: *.go
	GO111MODULE=on go build -ldflags '-X main.AyameVersion=${VERSION}' -o $(PROG)

build-darwin bd: *.go
	GO111MODULE=on GOOS=darwin GOARCH=amd64 go build -ldflags '-X main.AyameVersion=${VERSION}' -o bin/$(PROG)-darwin
build-linux bl: *.go
	GO111MODULE=on GOOS=linux GOARCH=amd64 go build -ldflags '-s -w -X main.AyameVersion=${VERSION}' -o bin/$(PROG)-linux

check:
	GO111MODULE=on go test ./...

fmt:
	go fmt ./...

clean:
	rm -rf $(PROG)

run r:
	./$(PROG)

kill k:
	pkill $(PROG)

log l:
	tail -f $(PROG).log
# ----------------------------------------------------------------------------------------
PROG_IMAGE=agilertc/$(PROG):$(VERSION)
PROG_NAME=$(PROG)
docker d:
	@echo "> make ([35mdocker[0m) [build|run|kill|ps] for [33m$(PROG_IMAGE)[0m"

docker-build db: $(PROG).go Dockerfile
	@-docker rmi $(PROG_IMAGE)
	@-PROG=$(PROG) docker build -t $(PROG_IMAGE) .
	@docker images $(PROG_IMAGE)

docker-run dr:
	@-docker run -d -p=3000:3000 -p=3443:3443 --name=$(PROG_NAME) $(PROG_IMAGE)
	@docker ps

# docker rm -f $(PROG_NAME)
docker-kill dk:
	@-docker stop $(PROG_NAME) | xargs docker rm
	@docker ps

docker-ps dp:
	@docker ps -f name=$(PROG_NAME)

docker-up du:
	@docker push $(PROG_IMAGE)
# ----------------------------------------------------------------------------------------
compose c:
	@echo "> make ([35mcompose[0m) [up|down] for $(PROG)"

compose-up cu:
	@docker-compose up -d

compose-down cd:
	@docker-compose down

compose-ps cp:
	@docker-compose ps
# ----------------------------------------------------------------------------------------
ngrok n:
	@echo "> make (ngrok) [install|run]"

ngrok-install ni:
	snap install ngrok

ngrok-run nr:
	ngrok http 3000
#-----------------------------------------------------------------------------------------
open o:
	@echo "> make (open) [orig|page|app]"

open-orig oo:
	xdg-open https://github.com/OpenAyame/ayame

open-page op:
	xdg-open https://github.com/sikang99/ayame

open-app oa:	# AppRTC
	xdg-open https://github.com/webrtc/apprtc
#-----------------------------------------------------------------------------------------
git g:
	@echo "> make (git) [update|login|tag|status]"

git-update gu:
	git add .
	git commit -m "start redis support"
	git push

git-login gl:
	git config --global user.email "sikang99@gmail.com"
	git config --global user.name "Stoney Kang"
	git config --global push.default matching
	git config credential.helper store

git-tag gt:
	git tag $(VERSION)
	git push --tags

git-status gs:
	git status
	git log --oneline -5
#-----------------------------------------------------------------------------------------
