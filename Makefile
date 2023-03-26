OS=""
ARCH=""

.PHONY: build
build:
	export GOPROXY="https://goproxy.io,direct"
	CGO_ENABLED=1 GOOS=${OS} GOARCH=${ARCH} go build -ldflags "-s -w" -o build/ethevent main.go

clean:
	rm -rf ./build

# swag 1.7.0
.PHONY: docs
docs:
	swag init -d . -g ./main.go -o ./docs