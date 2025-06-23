BINDIR ?=./bin

ifeq ($(origin TIMEWARRIORDB), undefined)
	ifneq ("$(wildcard $(HOME)/.timewarrior)","")
		TIMEWARRIORDB := $(HOME)/.timewarrior
	else ifneq ("$(wildcard $(HOME)/.config/timewarrior)","")
		TIMEWARRIORDB := $(HOME)/.config/timewarrior
	endif
endif
INSTALLDIR ?=$(HOME)/.local/bin

VERSION:=$(shell git rev-parse --abbrev-ref HEAD)-$(shell git rev-parse --short HEAD)

.PHONY: build

build:
	go build -ldflags="-X 'main.Version=$(VERSION)'" -o ${BINDIR}/twe ./cmd/twe/main.go
	go build -o ${BINDIR}/echo ./cmd/echo/main.go

clean:
	go clean -testcache
	rm -rf ./bin/*

install: clean build
	chmod +x ./bin/*
	mkdir -p $(INSTALLDIR)
	cp ./bin/twe $(INSTALLDIR)
	cp ./bin/echo $(TIMEWARRIORDB)/extensions/
	@echo "\nInstalled twe to ${INSTALLDIR}. Ensure you have added this directory to your PATH!"

uninstall: 
	rm -f $(TIMEWARRIORDB)/extensions/echo
	rm -f $(INSTALLDIR)/twe

test: 
	go test -v ./...
