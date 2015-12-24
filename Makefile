build:
	go build -o boltapi cmd/boltapi/main.go

install:
	go install github.com/marconi/boltapi/cmd/boltapi
