
GOPATH := $(shell which cygpath >/dev/null && cygpath -w $(CURDIR) || echo $(CURDIR))
export GOPATH

all: build

FILES := load/*.go ffnn/*.go config/*.go cmd/*/*.go

BASEPATH := github.com/cfstras/lfmnn

.PHONY: fix
fix:
	[ -f bin/goimports ] || make devdeps
	goimports -l -w $(FILES)

.PHONY: build
build: $(BASEPATH)/cmd/load \
		$(BASEPATH)/cmd/testnn \
		$(BASEPATH)/cmd/fmnn

.PHONY: $(BASEPATH)/cmd/%
$(BASEPATH)/cmd/%:
	go build -o bin/$(@F) $@

cmd/%: build $(BASEPATH)/cmd/%
	bin/$(@F)

.PHONY: clean
clean:
	rm -rf src/github.com/{shkh,skratchdot} src/code.google.com \
		pkg bin


.PHONY: deps
deps:
	go get github.com/shkh/lastfm-go/lastfm \
		github.com/skratchdot/open-golang/open

.PHONY: devdeps
devdeps: deps
	go get code.google.com/p/go.tools/cmd/goimports
