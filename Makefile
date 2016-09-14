build:
	go build -o ./bin/conftmpl ./cmd/conftmpl
	go build -o ./bin/confadm ./cmd/confadm

run: build
	./bin/conftmpl
