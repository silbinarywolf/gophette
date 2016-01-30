if not exist bin md bin
go build -ldflags -H=windowsgui -tags fatbin -o bin\gophette_exe.exe
copy resource\resources.blob bin\resources.blob
cd bin
fatbin -f.dir=. -f.exe=gophette_exe.exe -f.out=Gophette.exe
del gophette_exe.exe
del resources.blob
cd ..