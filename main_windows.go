package main

import (
	"bytes"
	"encoding/binary"
	"github.com/gonutz/blob"
	"github.com/gonutz/d3d9"
	"github.com/gonutz/mixer"
	"github.com/gonutz/mixer/wav"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/sdl_image"
	"image"
	"image/draw"
	"image/png"
	"os"
	"runtime"
	"time"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	// TODO enable VSync in D3D

	check(sdl.Init(0))
	defer sdl.Quit()

	check(mixer.Init())
	defer mixer.Close()

	if img.Init(img.INIT_PNG)&img.INIT_PNG == 0 {
		panic("error init png")
	}
	defer img.Quit()

	window, err := sdl.CreateWindow(
		"Gophette's Adventure",
		sdl.WINDOWPOS_CENTERED,
		sdl.WINDOWPOS_CENTERED,
		800,
		600,
		sdl.WINDOW_RESIZABLE,
	)
	check(err)
	defer window.Destroy()
	fullscreen := false

	info, err := window.GetWMInfo()
	winInfo := info.GetWindowsInfo()
	windowHandle := winInfo.Window

	sdl.ShowCursor(0)

	check(d3d9.Init())
	defer d3d9.Close()

	d3d, err := d3d9.Create(d3d9.SDK_VERSION)
	check(err)
	defer d3d.Release()

	maxScreenW, maxScreenH := 0, 0
	displayCount, err := sdl.GetNumVideoDisplays()
	check(err)
	for i := 0; i < displayCount; i++ {
		var mode sdl.DisplayMode
		err := sdl.GetCurrentDisplayMode(i, &mode)
		if err == nil {
			if int(mode.W) > maxScreenW {
				maxScreenW = int(mode.W)
			}
			if int(mode.H) > maxScreenH {
				maxScreenH = int(mode.H)
			}
		}
	}

	device, _, err := d3d.CreateDevice(
		d3d9.ADAPTER_DEFAULT,
		d3d9.DEVTYPE_HAL,
		windowHandle,
		d3d9.CREATE_HARDWARE_VERTEXPROCESSING,
		d3d9.PRESENT_PARAMETERS{
			BackBufferWidth:  uint(maxScreenW),
			BackBufferHeight: uint(maxScreenH),
			BackBufferFormat: d3d9.FMT_A8R8G8B8,
			BackBufferCount:  1,
			Windowed:         true,
			SwapEffect:       d3d9.SWAPEFFECT_DISCARD,
			HDeviceWindow:    windowHandle,
		},
	)
	check(err)
	defer device.Release()

	camera := newWindowCamera(window.GetSize())
	graphics := newWindowsGraphics(device, camera)

	assetLoader := newWindowsAssetLoader(device, graphics, camera)
	defer assetLoader.close()

	// charIndex selects which character is being controlled by the user, for
	// the final game this must be 0 but for creating the "AI" for Barney, set
	// this to 1 and delete the recorded inputs so they are not applied
	// additionally to the user controls

	var charIndex int
	const recordingAI = false // NOTE switch for development mode
	if !recordingAI {
		charIndex = 0
	} else {
		charIndex = 1
		recordedInputs = recordedInputs[:0]
		recordingInput = true
	}

	game := NewGame(
		assetLoader,
		graphics,
		camera,
		charIndex,
	)

	frameTime := time.Second / 65
	lastUpdate := time.Now().Add(-frameTime)

	// TODO bring back the music, ogg can not be loaded right now, maybe
	// convert the file to wav and play that
	//music, err := mix.LoadMUS("./rsc/background_music.ogg")
	//if err != nil {
	//	fmt.Println("error loading music:", err)
	//} else {
	//	defer music.Free()
	//	music.FadeIn(-1, 500)
	//}

	for game.Running() {
		for e := sdl.PollEvent(); e != nil; e = sdl.PollEvent() {
			switch event := e.(type) {
			case *sdl.KeyDownEvent:
				if event.Repeat == 0 {
					switch event.Keysym.Sym {
					case sdl.K_LEFT:
						game.HandleInput(InputEvent{GoLeft, true, charIndex})
					case sdl.K_RIGHT:
						game.HandleInput(InputEvent{GoRight, true, charIndex})
					case sdl.K_UP:
						game.HandleInput(InputEvent{Jump, true, charIndex})
					case sdl.K_ESCAPE:
						game.HandleInput(InputEvent{QuitGame, true, charIndex})
					}
				}
			case *sdl.KeyUpEvent:
				switch event.Keysym.Sym {
				case sdl.K_LEFT:
					game.HandleInput(InputEvent{GoLeft, false, charIndex})
				case sdl.K_RIGHT:
					game.HandleInput(InputEvent{GoRight, false, charIndex})
				case sdl.K_UP:
					game.HandleInput(InputEvent{Jump, false, charIndex})
				case sdl.K_F11:
					if fullscreen {
						window.SetFullscreen(0)
					} else {
						window.SetFullscreen(sdl.WINDOW_FULLSCREEN_DESKTOP)
					}
					fullscreen = !fullscreen
				case sdl.K_ESCAPE:
					game.HandleInput(InputEvent{QuitGame, false, charIndex})
				}
			case *sdl.WindowEvent:
				if event.Event == sdl.WINDOWEVENT_SIZE_CHANGED {
					width, height := int(event.Data1), int(event.Data2)
					camera.setWindowSize(width, height)
				}
			case *sdl.QuitEvent:
				game.HandleInput(InputEvent{QuitGame, true, charIndex})
			}
		}

		now := time.Now()
		dt := now.Sub(lastUpdate)
		if dt > frameTime {
			game.Update()
			lastUpdate = now
		}

		check(device.Clear(nil, d3d9.CLEAR_TARGET, d3d9.ColorRGB(0, 95, 83), 1, 0))
		game.Render()
		graphics.flush()
		check(device.Present(nil, nil, nil, nil))
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type wavSound struct {
	source mixer.SoundSource
}

func (s *wavSound) PlayOnce() {
	s.source.PlayOnce()
}

type d3dImage struct {
	camera *windowCamera
}

type textureImage struct {
	graphics      *windowsGraphics
	texture       d3d9.Texture
	width, height int
}

func (img *textureImage) DrawAt(x, y int) {
	// this call is referred to the graphics which will accumulate all calls
	// and then flush them out in one go at rendering time
	img.graphics.drawImageAt(img, x, y)
}

func (img *textureImage) Size() (int, int) {
	return img.width, img.height
}

func newWindowsAssetLoader(
	device d3d9.Device,
	graphics *windowsGraphics,
	camera *windowCamera,
) *windowsAssetloader {
	l := &windowsAssetloader{
		device:   device,
		graphics: graphics,
		camera:   camera,
		sounds:   make(map[string]*wavSound),
		images:   make(map[string]*textureImage),
	}
	check(l.loadResources())
	return l
}

type windowsAssetloader struct {
	device    d3d9.Device
	graphics  *windowsGraphics
	resources *blob.Blob
	camera    *windowCamera
	sounds    map[string]*wavSound
	images    map[string]*textureImage
}

func (l *windowsAssetloader) loadResources() error {
	rscFile, err := os.Open("./resource/resources.blob")
	if err != nil {
		return err
	}
	defer rscFile.Close()
	l.resources, err = blob.Read(rscFile)
	return err
}

func (l *windowsAssetloader) close() {
	for i := range l.images {
		l.images[i].texture.Release()
	}
}

func (l *windowsAssetloader) LoadImage(id string) Image {
	if img, ok := l.images[id]; ok {
		return img
	}

	data, _ := l.resources.GetByID(id)
	if data == nil {
		panic("unknown image resource: " + id)
	}
	ping, err := png.Decode(bytes.NewReader(data))
	check(err)

	var nrgba *image.NRGBA
	if asNRGBA, ok := ping.(*image.NRGBA); ok {
		nrgba = asNRGBA
	} else {
		nrgba = image.NewNRGBA(ping.Bounds())
		draw.Draw(nrgba, nrgba.Bounds(), ping, image.ZP, draw.Src)
	}

	texture, err := l.device.CreateTexture(
		uint(nrgba.Bounds().Dx()),
		uint(nrgba.Bounds().Dy()),
		1,
		d3d9.USAGE_SOFTWAREPROCESSING,
		d3d9.FMT_A8R8G8B8,
		d3d9.POOL_MANAGED,
		nil,
	)
	check(err)
	check(texture.LockedSetData(
		0,
		nil,
		d3d9.LOCK_DISCARD,
		nrgba.Pix,
		nrgba.Stride,
		nrgba.Bounds().Dy(),
	))

	img := &textureImage{
		l.graphics,
		texture,
		nrgba.Bounds().Dx(),
		nrgba.Bounds().Dy(),
	}
	l.images[id] = img

	return img
}

func (l *windowsAssetloader) LoadSound(id string) Sound {
	if sound, ok := l.sounds[id]; ok {
		return sound
	}
	data, _ := l.resources.GetByID(id)
	if data == nil {
		panic("unknown sound resource: " + id)
	}

	wave, err := wav.Load(bytes.NewReader(data))
	check(err)
	source, err := mixer.NewSoundSource(wave)
	check(err)
	sound := &wavSound{source}
	l.sounds[id] = sound

	return sound
}

func (l *windowsAssetloader) LoadRectangle(id string) Rectangle {
	data, found := l.resources.GetByID(id)
	if !found {
		panic("unknown rectangle resource: " + id)
	}
	reader := bytes.NewReader(data)
	var r rect
	check(binary.Read(reader, binary.LittleEndian, &r))
	return Rectangle{int(r.X), int(r.Y), int(r.W), int(r.H)}
}

type rect struct {
	X, Y, W, H int32
}

type windowsGraphics struct {
	device                   d3d9.Device
	camera                   *windowCamera
	textureVS                d3d9.VertexShader
	texturePS                d3d9.PixelShader
	vertexBuffer             d3d9.VertexBuffer
	vertexBufferLength       int
	textureCoordBuffer       d3d9.VertexBuffer
	textureCoordBufferLength int
	vertices                 []float32
	textureCoords            []float32
}

func newWindowsGraphics(device d3d9.Device, camera *windowCamera) *windowsGraphics {
	g := &windowsGraphics{
		device: device,
		camera: camera,
	}
	check(g.init())
	return g
}

func (g *windowsGraphics) init() error {
	textureVS, err := g.device.CreateVertexShaderFromBytes(dxTextureVso)
	if err != nil {
		return err
	}
	texturePS, err := g.device.CreatePixelShaderFromBytes(dxTexturePso)
	if err != nil {
		return err
	}
	g.textureVS = textureVS
	g.texturePS = texturePS

	return nil
}

func (graphics *windowsGraphics) FillRect(rect Rectangle, r, g, b, a uint8) {
	// TODO
	//check(graphics.renderer.SetDrawColor(r, g, b, a))
	//rect = rect.MoveBy(graphics.camera.offset())
	//sdlRect := sdl.Rect{int32(rect.X), int32(rect.Y), int32(rect.W), int32(rect.H)}
	//graphics.renderer.FillRect(&sdlRect)
}

func (graphics *windowsGraphics) ClearScreen(r, g, b uint8) {
	check(graphics.device.Clear(
		nil,
		d3d9.CLEAR_TARGET,
		d3d9.ColorRGB(r, g, b),
		1,
		0,
	))
}

func (g *windowsGraphics) drawImageAt(img *textureImage, x, y int) {

}

func (g *windowsGraphics) flush() {
	check(g.device.SetVertexShader(g.textureVS))
	check(g.device.SetPixelShader(g.texturePS))
}
