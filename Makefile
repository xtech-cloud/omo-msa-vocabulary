APP_NAME := xm-msa-vocabulary
BUILD_VERSION   := $(shell git tag --contains)
BUILD_TIME      := $(shell date "+%F %T")
COMMIT_SHA1     := $(shell git rev-parse HEAD )

.PHONY: build
build:
	export GOPROXY=https://goproxy.cn
	go build -ldflags \
		"\
		-X 'main.BuildVersion=${BUILD_VERSION}' \
		-X 'main.BuildTime=${BUILD_TIME}' \
		-X 'main.CommitID=${COMMIT_SHA1}' \
		"\
		-o ./bin/${APP_NAME}

.PHONY: run
run:
	./bin/${APP_NAME}

.PHONY: call
call:
	MICRO_REGISTRY=consul micro call omo.msa.vocabulary RelationService.GetAll '{"uid":""}'

.PHONY: tester
tester:
	go build -o ./bin/ ./tester

.PHONY: dist
dist:
	mkdir -p dist
	rm -f dist/${APP_NAME}-${BUILD_VERSION}.tar.gz
	tar -zcf dist/${APP_NAME}-${BUILD_VERSION}.tar.gz ./bin/${APP_NAME}

.PHONY: docker
docker:
	docker build . -t omo.msa.vocabulary:latest

.PHONY: updev
updev:
	scp -P 2209 dist/${APP_NAME}-${BUILD_VERSION}.tar.gz root@192.168.1.10:/root/

.PHONY: upload
upload:
	scp -P 9099 dist/${APP_NAME}-${BUILD_VERSION}.tar.gz root@47.93.209.105:/root/
