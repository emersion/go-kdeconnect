all:
	go build kdeconnect.go
start:
	go run kdeconnect.go

.PHONY: all start
