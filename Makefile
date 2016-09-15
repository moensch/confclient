build:
	go build -o ./bin/conftpl ./cmd/conftpl
	go build -o ./bin/confadm ./cmd/confadm

run: build
	./bin/conftpl
