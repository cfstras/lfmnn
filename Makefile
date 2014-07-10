
GOPATH := $(CURDIR)
export GOPATH

all: build

FILES := load/*.go ffnn/*.go cmd/*/*.go

.PHONY: fix
fix:
	[ -f bin/goimports ] || make devdeps
	goimports -l -w $(FILES)

.PHONY: build
build: github.com/cfstras/lfmnn/cmd/load-lastfm \
		github.com/cfstras/lfmnn/cmd/testnn

.PHONY: github.com/cfstras/lfmnn/cmd/%
github.com/cfstras/lfmnn/cmd/%:
	go build -o bin/$(@F) $@

cmd/%: build github.com/cfstras/lfmnn/cmd/%
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
