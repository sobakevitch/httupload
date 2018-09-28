APP?=httupload
PROJECT?=sovakevitch/httupload
GOOS?=linux
GOARCH?=amd64
CERTFILE?=server

all: linux windows

clean:
	/bin/rm -f $(APP) $(APP).exe

cleanall: clean
	/bin/rm -f $(CERTFILE).crt $(CERTFILE).key

linux:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(APP)

test:
	go test -v -race ./...

windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=$(GOARCH) go build -o $(APP).exe
	
.PHONY: cert
cert:
	@openssl req -x509 -newkey rsa:4096 -keyout $(CERTFILE).key -out $(CERTFILE).crt -days 365 -nodes -subj '/CN=localhost'

install: linux
	/bin/mv $(APP) $(GOPATH)/bin
