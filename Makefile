.PHONY: nuke
nuke:
	rm -rf tmp/* || true
	docker rmi chaseisabelle/gocrx:local || true

.PHONY: data
data:
	mkdir tmp || true
	docker run --rm -w /workdir -v $(PWD)/tmp:/workdir alpine/git \
		clone https://github.com/GoogleChrome/chrome-extensions-samples.git || true
	rm -rf tmp/test
	mkdir -p tmp/test
	cp -R tmp/chrome-extensions-samples/functional-samples/tutorial.hello-world/* tmp/test

.PHONY: run
run:
	docker compose run --rm ${service} ${command}

.PHONY: vet
vet:
	make run service=vetter

.PHONY: test
test:
	make data
	make run service=tester

.PHONY: cover
cover:
	make run service=coverer

.PHONY: lint
lint:
	make run service=linter

.PHONY: trufflehog
trufflehog:
	make run service=trufflehog