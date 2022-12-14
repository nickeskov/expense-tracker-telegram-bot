CURDIR=$(shell pwd)
BINDIR=${CURDIR}/bin
GOVER=$(shell go version | perl -nle '/(go\d\S+)/; print $$1;')
MOCKGEN=${BINDIR}/mockgen_${GOVER}
SMARTIMPORTS=${BINDIR}/smartimports_${GOVER}
LINTVER=v1.49.0
LINTBIN=${BINDIR}/lint_${GOVER}_${LINTVER}
PACKAGE=gitlab.ozon.dev/mr.eskov1/telegram-bot/cmd/bot

all: format build test lint

build: bindir
	go build -o ${BINDIR}/bot ${PACKAGE}

test:
	go test ./...

run:
	go run ${PACKAGE}

generate: install-mockgen
	#${MOCKGEN} -source=internal/expense/repository.go -destination=internal/generated/mocks/expense/repository.go # unused for now
	${MOCKGEN} -source=internal/expense/usecase.go -destination=internal/generated/mocks/expense/usecase.go
	${MOCKGEN} -source=internal/clients/tg/tgclient.go -destination=internal/generated/mocks/clients/tg.go
	${MOCKGEN} -source=internal/user/user.go -destination=internal/generated/mocks/user/user.go
	#${MOCKGEN} -source=internal/exrate/exrate.go -destination=internal/generated/mocks/exrate/exrate.go #unused for now
	#${MOCKGEN} -source=internal/providers/providers.go -destination=internal/generated/mocks/providers/providers.go #unused for now

lint: install-lint
	${LINTBIN} run

precommit: format build test lint
	echo "OK"

bindir:
	mkdir -p ${BINDIR}

format: install-smartimports
	${SMARTIMPORTS} -exclude internal/mocks

install-mockgen: bindir
	test -f ${MOCKGEN} || \
		(GOBIN=${BINDIR} go install github.com/golang/mock/mockgen@v1.6.0 && \
		mv ${BINDIR}/mockgen ${MOCKGEN})

install-lint: bindir
	test -f ${LINTBIN} || \
		(GOBIN=${BINDIR} go install github.com/golangci/golangci-lint/cmd/golangci-lint@${LINTVER} && \
		mv ${BINDIR}/golangci-lint ${LINTBIN})

install-smartimports: bindir
	test -f ${SMARTIMPORTS} || \
		(GOBIN=${BINDIR} go install github.com/pav5000/smartimports/cmd/smartimports@latest && \
		mv ${BINDIR}/smartimports ${SMARTIMPORTS})

docker-run:
	sudo docker compose up
