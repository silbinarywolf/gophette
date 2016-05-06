if not exist bin md bin
go build -tags sdl2 -ldflags -H=windowsgui -o bin\gophette.exe
go get -u github.com/gonutz/payload/cmd/payload
payload -exe=bin\gophette.exe -data=resource\resources.blob
