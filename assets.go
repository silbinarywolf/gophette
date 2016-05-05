package main

import "time"

type Graphics interface {
	ClearScreen(r, g, b uint8)
}

type Image interface {
	DrawAt(x, y int)
	Size() (width, height int)
}

type Sound interface {
	PlayOnce()
	Length() time.Duration
}

type AssetLoader interface {
	LoadImage(id string) Image
	LoadSound(id string) Sound
	LoadRectangle(id string) Rectangle
}
