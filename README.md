# Gophette's Great Adventure

Gophette, the only female Gopher in the Gocave, is out for a walk in the forest when she hears a strange voice...

Evil Doctor Barney Starsoup is sitting in his cabin, looking at the programming language news groups as he finds out about the nice little language that Gophette so admires.

![Barney Starsoup](https://raw.githubusercontent.com/gophergala2016/gophette/master/screenshots/barney_starsoup.png)
![Gophette](https://raw.githubusercontent.com/gophergala2016/gophette/master/screenshots/gophette.png)

Doctor Starsoup has a reputation of adding terrible features to perfectly fine languages and hence he seeks to find the secret Gocave and make it his own.

Can you beat Evil Doctor Barney Starsoup in a race to your home and warn the other Gophers about the threat before it is too late?

![Race](https://raw.githubusercontent.com/gophergala2016/gophette/master/screenshots/race.png)

Here is a video of the gameplay:
![Gameplay](https://github.com/gophergala2016/gophette/raw/master/screenshots/gameplay.flv)

# Build

**Note** that you need a C compiler to build this game on all operating systems.

##Windows

You need a C compiler so if you do not have one, install [MinGW](http://sourceforge.net/projects/mingw/files/).

On Windows the game uses DirectX per default. The dependencies are installed automatically when you get the game.
Run the following command:

	go get github.com/gonutz/gophette

and go into the source directory under %GOPATH%\src\github.com\gonutz\gophette. There run the Windows build script:

	build_win.bat

The resulting executable will be placed inside the gophette directory under bin\gophette.exe. It contains the resource data (sounds and images) and can be run on any Windows maching with Windows XP or later.

Note however that if you build the game with a 64 bit compiler, it will only run on 64 bit Windows, while 32 bit executables run on both 32 and 64 bit Windows.

##Linux

On Linux the game uses the SDL2 library, so make sure to install it by running:

	sudo apt-get install libsdl2-dev
	sudo apt-get install libsdl2-image-dev
	sudo apt-get install libsdl2-mixer-dev

After that you can get the game with:

	go get github.com/gonutz/gophette

Then go into the source directory under $GOPATH/src/github.com/gonutz/gophette. There run the Linux build script:

	./build_linux.sh

The resulting executable will be placed inside the gophette directory under bin/gophette.exe. It contains the resource data (sounds and images) and can be run from any directory.

##OS X

On OS X the game uses the SDL2 library, so make sure to install it by running:

	brew install sdl2
	brew install --with-libvorbis sdl2_mixer
	brew install sdl2_image

After that you can get the game with:

	go get github.com/gonutz/gophette

Then go into the source directory under $GOPATH/src/github.com/gonutz/gophette. There run the Linux build script:

	./build_linux.sh

The resulting executable will be placed inside the gophette directory under bin/gophette.exe. It contains the resource data (sounds and images) and can be run from any directory.

# About

I created this as a solo project, meaning this is all programmer art (graphics and sound). I have created small games in the past, first in C++ and now in Go.

I hope people enjoy this game and realize that Go is very capable of creating desktops apps.