.PHONY: test

test:
	go run tests/v1/main.go --token "$(token)" --url "$(url)" --port 3000

