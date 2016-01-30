#/!bin/bash
if [ ! -d bin ]; then
  mkdir bin
fi
go build -tags fatbin -o bin/gophette_exe
cp resource/resources.blob bin/resources.blob
cd bin
fatbin -f.dir=. -f.exe=gophette_exe -f.out=Gophette
rm gophette_exe
rm resources.blob
