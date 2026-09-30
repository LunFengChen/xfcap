.PHONY: test android clean

test:
	go test ./...

android:
	CGO_ENABLED=0 GOOS=android GOARCH=arm64 go build -trimpath -ldflags '-s -w' -o dist/xfcap ./cmd/xfcap

clean:
	rm -rf dist
