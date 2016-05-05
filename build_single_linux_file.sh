#/!bin/bash
if [ ! -d bin ]; then
  mkdir bin
fi
go build -o bin/gophette_exe
go get -u github.com/gonutz/payload/cmd/payload
payload -exe=bin/gophette_exe -data=resource/resources.blob -output=bin/gophette
rm bin/gophette_exe
