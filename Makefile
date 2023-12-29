.PHONY: test
test:
	mkdir tmp || true
	docker run --rm -w /workdir -v $(PWD)/tmp:/workdir alpine/git \
		clone https://github.com/GoogleChrome/chrome-extensions-samples.git || true
	rm -rf tmp/test
	mkdir -p tmp/test || true
	cp -R tmp/chrome-extensions-samples/functional-samples/tutorial.hello-world/* tmp/test
	docker build -t gocrx .
	docker run --rm -w /workdir -v $(PWD):/workdir gocrx go test -v ./...

