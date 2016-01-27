package main

import (
	"bytes"
	"encoding/binary"
	"github.com/disintegration/imaging"
	"github.com/gonutz/atlas"
	"github.com/gonutz/blob"
	"github.com/gonutz/xcf"
	"github.com/nfnt/resize"
	"image"
	"image/png"
	"io/ioutil"
	"os"
)

const scale = 0.33

var byteOrder = binary.LittleEndian

func main() {
	resources := blob.New()
	textureAtlas := atlas.New(2048)

	gophette, err := xcf.LoadFromFile("./gophette.xcf")
	check(err)
	barney, err := xcf.LoadFromFile("./barney.xcf")
	check(err)

	// create the collision information for Gophette and Barney
	addCollisionInfo := func(canvas xcf.Canvas, id string) {
		collision := canvas.GetLayerByName("collision")
		left, top := findTopLeftNonTransparentPixel(collision)
		right, bottom := findBottomRightNonTransparentPixel(collision)
		// scale the collision rect just like the images
		left = int(0.5 + scale*float64(left))
		top = int(0.5 + scale*float64(top))
		right = int(0.5 + scale*float64(right))
		bottom = int(0.5 + scale*float64(bottom))
		width, height := right-left+1, bottom-top+1
		r := rect{int32(left), int32(top), int32(width), int32(height)}
		buffer := bytes.NewBuffer(nil)
		check(binary.Write(buffer, byteOrder, &r))
		resources.Append(id, buffer.Bytes())
	}
	addCollisionInfo(gophette, "hero collision")
	addCollisionInfo(barney, "barney collision")

	addImage := func(img image.Image, id string) {
		_, err := textureAtlas.Add(id, img)
		check(err)
	}

	// create the image resources
	for _, layer := range []string{
		"jump",
		"run1",
		"run2",
		"run3",
	} {
		small := scaleImage(gophette.GetLayerByName(layer))
		addImage(small, "gophette_left_"+layer)
		addImage(imaging.FlipH(small), "gophette_right_"+layer)
	}

	for _, layer := range []string{
		"stand",
		"jump",
		"run1",
		"run2",
		"run3",
		"run4",
		"run5",
		"run6",
	} {
		smallLeft := scaleImage(barney.GetLayerByName("left_" + layer))
		smallRight := scaleImage(barney.GetLayerByName("right_" + layer))
		addImage(smallLeft, "barney_left_"+layer)
		addImage(smallRight, "barney_right_"+layer)
	}

	grass, err := xcf.LoadFromFile("./grass.xcf")
	check(err)
	for _, layer := range []string{
		"grass left",
		"grass right",
		"grass center 1",
		"grass center 2",
		"grass center 3",
	} {
		addImage(grass.GetLayerByName(layer), layer)
	}

	grassLong, err := xcf.LoadFromFile("./grass_long.xcf")
	check(err)
	for _, layer := range []string{
		"grass long 1",
		"grass long 2",
		"grass long 3",
	} {
		addImage(grassLong.GetLayerByName(layer), layer)
	}

	ground, err := xcf.LoadFromFile("./ground.xcf")
	check(err)
	for _, layer := range []string{
		"ground left",
		"ground right",
		"ground center 1",
		"ground center 2",
		"ground center 3",
	} {
		addImage(ground.GetLayerByName(layer), layer)
	}

	groundLong, err := xcf.LoadFromFile("./ground_long.xcf")
	check(err)
	for _, layer := range []string{
		"ground long 1",
		"ground long 2",
	} {
		addImage(groundLong.GetLayerByName(layer), layer)
	}

	rock, err := xcf.LoadFromFile("./rock.xcf")
	check(err)
	addImage(scaleImage(rock.GetLayerByName("rock")), "square rock")

	tree, err := xcf.LoadFromFile("./tree.xcf")
	check(err)
	smallTree := scaleImage(tree.GetLayerByName("small"))
	addImage(smallTree, "small tree")

	tree, err = xcf.LoadFromFile("./tree_big.xcf")
	check(err)
	bigTree := scaleImage(tree.GetLayerByName("big"))
	addImage(bigTree, "big tree")

	tree, err = xcf.LoadFromFile("./tree_huge.xcf")
	check(err)
	hugeTree := scaleImage(tree.GetLayerByName("huge"))
	addImage(hugeTree, "huge tree")

	cave, err := xcf.LoadFromFile("./cave.xcf")
	check(err)
	addImage(scaleImage(cave.GetLayerByName("cave back")), "cave back")
	addImage(scaleImage(cave.GetLayerByName("cave front")), "cave front")

	intro, err := xcf.LoadFromFile("./intro.xcf")
	check(err)
	addImage(scaleImageToFactor(intro.GetLayerByName("pc 1"), 0.67), "intro pc 1")
	addImage(scaleImageToFactor(intro.GetLayerByName("pc 2"), 0.67), "intro pc 2")
	addImage(scaleImageToFactor(intro.GetLayerByName("gophette"), 0.67), "intro gophette")

	music, err := ioutil.ReadFile("./background_music.ogg")
	check(err)
	resources.Append("music", music)

	for _, sound := range []string{
		"win",
		"lose",
		"fall",
		"barney wins",
		"barney intro text",
		"whistle",
		"instructions",
	} {
		data, err := ioutil.ReadFile(sound + ".wav")
		check(err)
		resources.Append(sound, data)
	}

	resources.Append("atlas", imageToBytes(textureAtlas))
	for _, sub := range textureAtlas.SubImages {
		resources.Append(
			sub.ID,
			toRectData(sub.Bounds().Sub(textureAtlas.Bounds().Min)),
		)
	}

	resourceFile, err := os.Create("../resource/resources.blob")
	check(err)
	defer resourceFile.Close()
	resources.Write(resourceFile)
}

func imageToBytes(img image.Image) []byte {
	buffer := bytes.NewBuffer(nil)
	check(png.Encode(buffer, img))
	return buffer.Bytes()
}

func findTopLeftNonTransparentPixel(img image.Image) (x, y int) {
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a != 0 {
				return x, y
			}
		}
	}
	return -1, -1
}

func findBottomRightNonTransparentPixel(img image.Image) (x, y int) {
	for y := img.Bounds().Max.Y - 1; y >= img.Bounds().Min.Y; y-- {
		for x := img.Bounds().Max.X - 1; x >= img.Bounds().Min.X; x-- {
			_, _, _, a := img.At(x, y).RGBA()
			if a != 0 {
				return x, y
			}
		}
	}
	return -1, -1
}

func scaleImage(img image.Image) image.Image {
	return scaleImageToFactor(img, scale)
}

func scaleImageToFactor(img image.Image, f float64) image.Image {
	return resize.Resize(
		uint(0.5+f*float64(img.Bounds().Dx())),
		uint(0.5+f*float64(img.Bounds().Dy())),
		img,
		resize.Bicubic,
	)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type rect struct {
	X, Y, W, H int32
}

func toRectData(bounds image.Rectangle) []byte {
	buf := bytes.NewBuffer(nil)
	r := rect{
		int32(bounds.Min.X),
		int32(bounds.Min.Y),
		int32(bounds.Dx()),
		int32(bounds.Dy()),
	}
	check(binary.Write(buf, byteOrder, &r))
	return buf.Bytes()
}
