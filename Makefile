SHELL := /bin/bash
TARGETS = lloyd-map lloyd-permute lloyd-uniq
PROJECT = lloyd

# http://docs.travis-ci.com/user/languages/go/#Default-Test-Script
test:
	go test -v ./...

bench:
	go test -bench=.

imports:
	goimports -w .

fmt:
	go fmt ./...

vet:
	go vet ./...

all: fmt test $(TARGETS)


install:
	go install

clean:
	go clean
	rm -f coverage.out
	rm -f $(TARGETS)
	rm -f $(PROJECT)-*.x86_64.rpm
	rm -f packaging/deb/$(PROJECT)*.deb
	rm -f $(PROJECT)*.deb
	rm -rf packaging/deb/$(PROJECT)/usr

cover:
	go get -d && go test -v	-coverprofile=coverage.out
	go tool cover -html=coverage.out

lloyd-map:
	go build -o lloyd-map cmd/lloyd-map/main.go

lloyd-permute:
	go build -o lloyd-permute cmd/lloyd-permute/main.go

lloyd-uniq:
	go build -o lloyd-uniq cmd/lloyd-uniq/main.go

# ==== packaging

deb: $(TARGETS)
	mkdir -p packaging/deb/$(PROJECT)/usr/sbin
	cp $(TARGETS) packaging/deb/$(PROJECT)/usr/sbin
	cd packaging/deb && fakeroot dpkg-deb --build $(PROJECT) .
	mv packaging/deb/*.deb .

rpm: $(TARGETS)
	mkdir -p $(HOME)/rpmbuild/{BUILD,SOURCES,SPECS,RPMS}
	cp ./packaging/rpm/$(PROJECT).spec $(HOME)/rpmbuild/SPECS
	cp $(TARGETS) $(HOME)/rpmbuild/BUILD
	./packaging/rpm/buildrpm.sh $(PROJECT)
	cp $(HOME)/rpmbuild/RPMS/x86_64/$(PROJECT)*.rpm .
