NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m
APP=scale
REVISION=$(shell git rev-parse --short HEAD)
BASE_VERSION=$(shell cat VERSION)
VERSION=$(BASE_VERSION)-$(REVISION)

all: build

build:
	@echo "$(OK_COLOR)==> Building revision $(VERSION)...$(NO_COLOR)"
	@script/build $(APP) $(VERSION)

docker:
	@echo "$(OK_COLOR)==> Building revision $(VERSION)...$(NO_COLOR)"
	@script/docker $(APP) $(VERSION)

docker.push: docker
	@echo "$(OK_COLOR)==> Building revision $(VERSION)...$(NO_COLOR)"
	@docker push scaleci/scale:$(VERSION)

format:
	go fmt ./...

test:
	@echo "$(OK_COLOR)==> Testing...$(NO_COLOR)"
	@script/test $(TEST)

release:
	@script/release $(VERSION)

install_equinox:
	@script/install_equinox $(VERSION)

.PHONY: all build clean test release install_equinox test prod
