
SHELL := /bin/bash

TAG=1.0.15
IMAGE=alicloud_monitoring:${TAG}

.PHONY: build
build:
	@docker build -t ${IMAGE} .

.PHONY: publish
publish:
	@docker push ${IMAGE}
