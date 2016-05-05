if not exist bin md bin
go build -tags sdl2 -ldflags -H=windowsgui -o bin\gophette_exe.exe
go get -u github.com/gonutz/payload/cmd/payload
payload -exe=bin\gophette_exe.exe -data=resource\resources.blob -output=bin\gophette.exe
del bin\gophette_exe.exe