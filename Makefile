REGION ?= eu-west-2
BUCKET ?= pubclub-artifacts
VERSION ?= 1.0.0

.PHONY: bucket
bucket:
	@if ! aws s3api head-bucket --bucket $(BUCKET) 2>/dev/null; then\
		aws s3api create-bucket --bucket $(BUCKET) --region $(REGION)\
		--create-bucket-configuration LocationConstraint=$(REGION) 1>/dev/null;\
	fi;

.PHONY: builds_directory
builds_directory:
	@if [ ! -d builds ]; then\
		mkdir -p builds;\
	fi;

.PHONY: build
build: builds_directory
	@cd cmd; \
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ../builds/main main.go; \
	cd ../builds/;\
	zip main.zip main

.PHONY: deploy
deploy: bucket build
	@aws s3 cp builds/main.zip s3://$(BUCKET)/ratings/v$(VERSION)/main.zip
