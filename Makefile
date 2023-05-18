.PHONY: build-windows

build-windows:
	GOOS=windows GOARCH=amd64 go build -o random.exe
