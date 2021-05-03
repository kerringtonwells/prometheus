build:
	go install -v .

test:
	go test -v ./...

image:
	docker build -t kalexiwells/slirunner:1.78.0 .
