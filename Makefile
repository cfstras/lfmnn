
GOPATH := $(shell which cygpath >/dev/null && cygpath -w $(CURDIR) || echo $(CURDIR))
export GOPATH

all: build

FILES := load/*.go ffnn/*.go config/*.go cmd/*/*.go

BASEFOLDER := github.com/cfstras
BASEPATH := $(BASEFOLDER)/lfmnn

src/$(BASEPATH):
	mkdir -p src/$(BASEFOLDER)
	[ "$(OS)" = "Windows_NT" ] && mklink /d src/$(BASEPATH) ..\..\..\
	[ "$(OS)" = "Windows_NT" ] || ln -s ../../../ src/$(BASEPATH)

.PHONY: build
build: $(BASEPATH)/cmd/load \
		$(BASEPATH)/cmd/testnn \
		$(BASEPATH)/cmd/fmnn

.PHONY: $(BASEPATH)/cmd/%
$(BASEPATH)/cmd/%: src/$(BASEPATH)
	go build -o bin/$(@F) $@

cmd/%: $(BASEPATH)/cmd/%
	bin/$(@F)

.PHONY: clean
clean:
	rm -rf src pkg bin

.PHONY: deps
deps:
	go get github.com/shkh/lastfm-go/lastfm \
		github.com/skratchdot/open-golang/open

.PHONY: fix
fix:
	[ -f bin/goimports ] || make devdeps
	go tool vet || make devdeps
	bin/goimports -l -w $(FILES)
	#TODO go vet

.PHONY: devdeps
devdeps: deps
	go get code.google.com/p/go.tools/cmd/goimports
	go get code.google.com/p/go.tools/cmd/vet
