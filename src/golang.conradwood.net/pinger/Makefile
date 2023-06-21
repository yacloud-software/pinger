.PHONY:	client server pinger
all:	client server pinger
pinger:
	cd pingexe && go install ${LDFLAGS} pinger.go `ls -1 *.go|grep -v pinger.go`
client:
	cd client && go install ${LDFLAGS} pinger-client.go
server:
	cd server && go install ${LDFLAGS} pinger-server.go `ls -1 *.go|grep -v pinger-server.go`
