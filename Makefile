.PHONY: data
data:
	mkdir tmp || true
	docker run --rm -w /workdir -v $(PWD)/tmp:/workdir alpine/git \
		clone https://github.com/GoogleChrome/chrome-extensions-samples.git || true
	rm -rf tmp/test
	mkdir -p tmp/test
	cp -R tmp/chrome-extensions-samples/functional-samples/tutorial.hello-world/* tmp/test
	rm -rf tmp/chrome-extensions-samples || true

.PHONY: test
test:
	make data
	docker build -t gocrx .
	docker run --rm -w /workdir -v $(PWD):/workdir gocrx go test -v ./...

