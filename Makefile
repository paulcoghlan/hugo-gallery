setup:
	mkdir -p ./test/sample-site/content/gallery
	mkdir -p ./test/sample-site/assets/images

build:
	go build

test: setup build
	HUGO_DIR=./test/sample-site ./hugo-gallery ./test/source/a/b/c gallery/a/b/c "test gallery" 

clean:
	rm -rf ./test/sample-site

.PHONY: test	