.PHONY: test

test:
	go run tests/v1/main.go --token "$(token)" --url "$(url)" --port 3000

# Steps
# 1. Boot up sdk test server on localhost:3000.
# 2. Wait a few seconds.
# 3. Run unit tests, which boots up mock server on localhost:4000.
# 4. etc.


