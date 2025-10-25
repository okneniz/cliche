default: test-report lint

test-report:
	go run gotest.tools/gotestsum@latest --format standard-verbose

test:
	go test --count=5 -coverprofile=coverage.out ./...
	# go test -v -count 1 -timeout 60s -coverprofile=coverage.out ./...

install-linter:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2

lint: fmt json_fmt
	golangci-lint run ./...

json_fmt:
	bash -c "for file in $(find ./testdata/onigmo -name '*.json'); do jq -M -e . < $file > $file.out && mv $file.out $file; done"
	bash -c "for file in $(find ./testdata/re2 -name '*.json'); do jq -M -e . < $file > $file.out && mv $file.out $file; done"

fmt:
	gofmt -w -s .

benchmark:
	# go test -v -bench=. -benchmem -memprofile memprofile.out -cpuprofile profile.out -count=3 -run=^# ./hash-map/...
	go test -v -bench=. -benchmem -count=5 -run=^# ./...

coverage:
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out

build:
	go build ./...

pub:
	GOPROXY=https://proxy.golang.org GO111MODULE=on go get github.com/okneniz/cliche
