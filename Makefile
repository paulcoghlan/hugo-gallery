build:
	go build

test: build
	go test 

clean:
	rm -rf ./test/sample-site

.PHONY: test clean	