// +build !sdl2

package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"image/draw"
	"image/png"
	"runtime"
	"syscall"
	"time"

	"github.com/gonutz/blob"
	"github.com/gonutz/d3d9"
	"github.com/gonutz/d3dmath"
	"github.com/gonutz/mixer"
	"github.com/gonutz/mixer/wav"
	"github.com/gonutz/payload"
	"github.com/gonutz/w32"
)

func init() {
	runtime.LockOSThread()
}

var (
	windowW           = 800
	windowH           = 800
	game              *Game
	camera            *windowCamera
	charIndex         int
	previousPlacement w32.WINDOWPLACEMENT
)

func toggleFullscreen(window w32.HWND) {
	style := w32.GetWindowLong(window, w32.GWL_STYLE)
	if style&w32.WS_OVERLAPPEDWINDOW != 0 {
		// go into full-screen
		var monitorInfo w32.MONITORINFO
		monitor := w32.MonitorFromWindow(window, w32.MONITOR_DEFAULTTOPRIMARY)
		if w32.GetWindowPlacement(window, &previousPlacement) &&
			w32.GetMonitorInfo(monitor, &monitorInfo) {
			w32.SetWindowLong(
				window,
				w32.GWL_STYLE,
				uint32(style & ^w32.WS_OVERLAPPEDWINDOW),
			)
			w32.SetWindowPos(
				window,
				0,
				int(monitorInfo.RcMonitor.Left),
				int(monitorInfo.RcMonitor.Top),
				int(monitorInfo.RcMonitor.Right-monitorInfo.RcMonitor.Left),
				int(monitorInfo.RcMonitor.Bottom-monitorInfo.RcMonitor.Top),
				w32.SWP_NOOWNERZORDER|w32.SWP_FRAMECHANGED,
			)
		}
		w32.ShowCursor(false)
	} else {
		// go into windowed mode
		w32.SetWindowLong(
			window,
			w32.GWL_STYLE,
			uint32(style|w32.WS_OVERLAPPEDWINDOW),
		)
		w32.SetWindowPlacement(window, &previousPlacement)
		w32.SetWindowPos(window, 0, 0, 0, 0, 0,
			w32.SWP_NOMOVE|w32.SWP_NOSIZE|w32.SWP_NOZORDER|
				w32.SWP_NOOWNERZORDER|w32.SWP_FRAMECHANGED,
		)
		w32.ShowCursor(true)
	}
}

func lowWord(x uint) int {
	return int(x & 0xFFFF)
}

func highWord(x uint) int {
	return int((x >> 16) & 0xFFFF)
}

func isKeyRepeat(l uintptr) bool {
	return l&(1<<30) != 0
}

func handleEvent(window w32.HWND, message uint32, w, l uintptr) uintptr {
	switch message {
	case w32.WM_KEYDOWN:
		if !isKeyRepeat(l) {
			switch w {
			case w32.VK_LEFT:
				game.HandleInput(InputEvent{GoLeft, true, charIndex})
			case w32.VK_RIGHT:
				game.HandleInput(InputEvent{GoRight, true, charIndex})
			case w32.VK_UP, w32.VK_SPACE:
				game.HandleInput(InputEvent{Jump, true, charIndex})
			case w32.VK_ESCAPE:
				game.HandleInput(InputEvent{QuitGame, true, charIndex})
			}
		}
		return 1
	case w32.WM_KEYUP:
		switch w {
		case w32.VK_LEFT:
			game.HandleInput(InputEvent{GoLeft, false, charIndex})
		case w32.VK_RIGHT:
			game.HandleInput(InputEvent{GoRight, false, charIndex})
		case w32.VK_UP, w32.VK_SPACE:
			game.HandleInput(InputEvent{Jump, false, charIndex})
		case w32.VK_F11:
			toggleFullscreen(window)
		case w32.VK_ESCAPE:
			game.HandleInput(InputEvent{QuitGame, false, charIndex})
			w32.PostQuitMessage(0)
		}
		return 1
	case w32.WM_SIZE:
		if camera != nil {
			windowW, windowH = lowWord(uint(l)), highWord(uint(l))
			camera.setWindowSize(windowW, windowH)
		}
		return 1
	case w32.WM_DESTROY:
		w32.PostQuitMessage(0)
		return 1
	default:
		return w32.DefWindowProc(window, message, w, l)
	}
}

func main() {
	windowHandle, err := openWindow("class name", handleEvent, 0, 0, windowW, windowH)
	check(err)
	window := w32.HWND(windowHandle)

	w32.SetWindowText(window, "Gophette's Adventure")

	check(mixer.Init())
	defer mixer.Close()

	w32.ShowCursor(false)
	defer w32.ShowCursor(true)

	d3d, err := d3d9.Create(d3d9.SDK_VERSION)
	check(err)
	defer d3d.Release()

	var maxScreenW, maxScreenH uint32
	for i := uint(0); i < d3d.GetAdapterCount(); i++ {
		mode, err := d3d.GetAdapterDisplayMode(i)
		if err == nil {
			if mode.Width > maxScreenW {
				maxScreenW = mode.Width
			}
			if mode.Height > maxScreenH {
				maxScreenH = mode.Height
			}
		}
	}
	if maxScreenW == 0 || maxScreenH == 0 {
		panic("no monitor detected")
	}

	device, _, err := d3d.CreateDevice(
		d3d9.ADAPTER_DEFAULT,
		d3d9.DEVTYPE_HAL,
		d3d9.HWND(windowHandle),
		d3d9.CREATE_HARDWARE_VERTEXPROCESSING,
		d3d9.PRESENT_PARAMETERS{
			BackBufferWidth:      maxScreenW,
			BackBufferHeight:     maxScreenH,
			BackBufferFormat:     d3d9.FMT_A8R8G8B8,
			BackBufferCount:      1,
			PresentationInterval: d3d9.PRESENT_INTERVAL_ONE, // enable VSync
			Windowed:             1,
			SwapEffect:           d3d9.SWAPEFFECT_DISCARD,
			HDeviceWindow:        d3d9.HWND(windowHandle),
		},
	)
	check(err)
	defer device.Release()

	check(device.SetRenderState(d3d9.RS_CULLMODE, uint32(d3d9.CULL_CW)))
	check(device.SetRenderState(d3d9.RS_SRCBLEND, d3d9.BLEND_SRCALPHA))
	check(device.SetRenderState(d3d9.RS_DESTBLEND, d3d9.BLEND_INVSRCALPHA))
	check(device.SetRenderState(d3d9.RS_ALPHABLENDENABLE, 1))

	camera = newWindowCamera(windowW, windowH)
	graphics := newWindowsGraphics(device, camera)
	defer graphics.close()

	assetLoader := newWindowsAssetLoader(device, graphics, camera)
	defer assetLoader.close()

	// charIndex selects which character is being controlled by the user, for
	// the final game this must be 0 but for creating the "AI" for Barney, set
	// this to 1 and delete the recorded inputs so they are not applied
	// additionally to the user controls

	const recordingAI = false // NOTE switch for development mode
	if !recordingAI {
		charIndex = 0
	} else {
		charIndex = 1
		recordedInputs = recordedInputs[:0]
		recordingInput = true
	}

	game = NewGame(
		assetLoader,
		graphics,
		camera,
		charIndex,
	)

	music := assetLoader.LoadSound("music_wav")
	go func() {
		for {
			music.PlayOnce()
			time.Sleep(music.Length() + 5*time.Second)
		}
	}()

	toggleFullscreen(window)

	frameTime := time.Second / 65
	lastUpdate := time.Now().Add(-frameTime)

	var msg w32.MSG
	w32.PeekMessage(&msg, 0, 0, 0, w32.PM_NOREMOVE)
	for msg.Message != w32.WM_QUIT {
		if w32.PeekMessage(&msg, 0, 0, 0, w32.PM_REMOVE) {
			w32.TranslateMessage(&msg)
			w32.DispatchMessage(&msg)
		} else {
			now := time.Now()
			dt := now.Sub(lastUpdate)
			if dt > frameTime {
				game.Update()
				lastUpdate = now
			}

			check(device.SetViewport(
				d3d9.VIEWPORT{0, 0, uint32(windowW), uint32(windowH), 0, 1},
			))
			check(device.Clear(
				nil,
				d3d9.CLEAR_TARGET,
				d3d9.ColorRGB(0, 95, 83),
				1,
				0,
			))
			game.Render()
			graphics.flush()
			check(device.Present(
				&d3d9.RECT{0, 0, int32(windowW), int32(windowH)},
				nil,
				0,
				nil,
			))
		}
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type messageCallback func(window w32.HWND, msg uint32, w, l uintptr) uintptr

func openWindow(
	className string,
	callback messageCallback,
	x, y, width, height int,
) (w32.HWND, error) {
	windowProc := syscall.NewCallback(callback)

	class := w32.WNDCLASSEX{
		WndProc:   windowProc,
		Cursor:    w32.LoadCursor(0, w32.MakeIntResource(w32.IDC_ARROW)),
		ClassName: syscall.StringToUTF16Ptr(className),
	}
	atom := w32.RegisterClassEx(&class)
	if atom == 0 {
		return 0, errors.New("RegisterClassEx failed")
	}

	window := w32.CreateWindowEx(
		0,
		syscall.StringToUTF16Ptr(className),
		nil,
		w32.WS_OVERLAPPEDWINDOW|w32.WS_VISIBLE,
		x, y, width, height,
		0, 0, 0, nil,
	)
	if window == 0 {
		return 0, errors.New("CreateWindowEx failed")
	}

	return window, nil
}

type wavSound struct {
	source mixer.SoundSource
}

func (s *wavSound) PlayOnce() {
	s.source.PlayOnce()
}

func (s *wavSound) Length() time.Duration {
	return s.source.Length()
}

type d3dImage struct {
	camera *windowCamera
}

type textureImage struct {
	graphics *windowsGraphics
	texture  *d3d9.Texture
	// these are the uv coordinates of the image in the texture
	u0, u1, v0, v1 float32
	width, height  int
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
	device *d3d9.Device,
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
	l.loadResources()
	l.graphics.textureAtlas = l.textureAtlas
	return l
}

type windowsAssetloader struct {
	device             *d3d9.Device
	graphics           *windowsGraphics
	resources          *blob.Blob
	camera             *windowCamera
	sounds             map[string]*wavSound
	images             map[string]*textureImage
	textureAtlas       *d3d9.Texture
	textureAtlasBounds image.Rectangle
}

func (l *windowsAssetloader) loadResources() {
	resourceData, err := payload.Read()
	check(err)
	l.resources, err = blob.Read(bytes.NewBuffer(resourceData))

	// load the texture atlas
	atlas, found := l.resources.GetByID("atlas")
	if !found {
		panic("texture atlas not found in resources")
	}

	ping, err := png.Decode(bytes.NewReader(atlas))
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
		0,
	)
	check(err)
	lockedRect, err := texture.LockRect(0, nil, d3d9.LOCK_DISCARD)
	check(err)
	lockedRect.SetAllBytes(nrgba.Pix, nrgba.Stride)
	check(texture.UnlockRect(0))

	l.textureAtlas = texture
	l.textureAtlasBounds = nrgba.Bounds()
}

func (l *windowsAssetloader) close() {
	for i := range l.images {
		l.images[i].texture.Release()
	}
	l.textureAtlas.Release()
}

func (l *windowsAssetloader) LoadImage(id string) Image {
	if img, ok := l.images[id]; ok {
		return img
	}

	data, _ := l.resources.GetByID(id)
	if data == nil {
		panic("unknown image resource: " + id)
	}

	var bounds rect
	check(binary.Read(bytes.NewReader(data), binary.LittleEndian, &bounds))

	var pixelW float32 = 1.0 / float32(l.textureAtlasBounds.Dx())
	var pixelH float32 = 1.0 / float32(l.textureAtlasBounds.Dy())
	left := float32(bounds.X) * pixelW
	bottom := float32(bounds.Y) * pixelH
	right := float32(bounds.X+bounds.W) * pixelW
	top := float32(bounds.Y+bounds.H) * pixelH

	img := &textureImage{
		l.graphics,
		l.textureAtlas,
		left, right, top, bottom,
		int(bounds.W),
		int(bounds.H),
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

	wave, err := wav.Read(bytes.NewReader(data))
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
	device                   *d3d9.Device
	textureAtlas             *d3d9.Texture
	camera                   *windowCamera
	textureVS                *d3d9.VertexShader
	texturePS                *d3d9.PixelShader
	vertexBuffer             *d3d9.VertexBuffer
	vertexBufferLength       int
	textureCoordBuffer       *d3d9.VertexBuffer
	textureCoordBufferLength int
	vertices                 []float32
	textureCoords            []float32
	vertexDecl               *d3d9.VertexDeclaration
}

func newWindowsGraphics(device *d3d9.Device, camera *windowCamera) *windowsGraphics {
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

	decl, err := g.device.CreateVertexDeclaration([]d3d9.VERTEXELEMENT{
		{0, 0, d3d9.DECLTYPE_FLOAT2, d3d9.DECLMETHOD_DEFAULT, d3d9.DECLUSAGE_POSITION, 0},
		{1, 0, d3d9.DECLTYPE_FLOAT2, d3d9.DECLMETHOD_DEFAULT, d3d9.DECLUSAGE_TEXCOORD, 0},
		d3d9.DeclEnd(),
	})
	check(err)
	g.vertexDecl = decl

	return nil
}

func (g *windowsGraphics) close() {
	g.textureCoordBuffer.Release()
	g.vertexBuffer.Release()
	g.vertexDecl.Release()
	g.texturePS.Release()
	g.textureVS.Release()
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
	dx, dy := g.camera.offset()
	x += dx
	y += dy

	xf, yf := float32(x), float32(y)
	// offset the pixels to correctly show the undistorted texture image, see
	// https://msdn.microsoft.com/en-us/library/windows/desktop/bb219690%28v=vs.85%29.aspx
	// NOTE this only works if the backbuffer has the same size as the window
	const pixOffset = -0.5
	g.vertices = append(g.vertices,
		xf+pixOffset, yf+pixOffset,
		xf+pixOffset, yf+float32(img.height)+pixOffset,
		xf+float32(img.width)+pixOffset, yf+pixOffset,
		xf+float32(img.width)+pixOffset, yf+pixOffset,
		xf+pixOffset, yf+float32(img.height)+pixOffset,
		xf+float32(img.width)+pixOffset, yf+float32(img.height)+pixOffset,
	)

	g.textureCoords = append(g.textureCoords,
		img.u0, img.v1,
		img.u0, img.v0,
		img.u1, img.v1,
		img.u1, img.v1,
		img.u0, img.v0,
		img.u1, img.v0,
	)
}

func (g *windowsGraphics) flush() {
	if len(g.vertices) == 0 {
		// nothing to do in this case
		return
	}

	if g.vertexBufferLength < len(g.vertices)*4 {
		if g.vertexBufferLength > 0 {
			g.vertexBuffer.Release()
		}
		var err error
		g.vertexBuffer, err = g.device.CreateVertexBuffer(
			uint(len(g.vertices)*4),
			d3d9.USAGE_WRITEONLY,
			0,
			d3d9.POOL_MANAGED,
			0,
		)
		check(err)
		g.vertexBufferLength = len(g.vertices) * 4
	}
	vbMem, err := g.vertexBuffer.Lock(0, 0, d3d9.LOCK_DISCARD)
	check(err)
	vbMem.SetFloat32s(0, g.vertices)
	check(g.vertexBuffer.Unlock())

	if g.textureCoordBufferLength < len(g.textureCoords)*4 {
		if g.textureCoordBufferLength > 0 {
			g.textureCoordBuffer.Release()
		}
		var err error
		g.textureCoordBuffer, err = g.device.CreateVertexBuffer(
			uint(len(g.textureCoords)*4),
			d3d9.USAGE_WRITEONLY,
			0,
			d3d9.POOL_MANAGED,
			0,
		)
		check(err)
		g.textureCoordBufferLength = len(g.textureCoords) * 4
	}
	texMem, err := g.textureCoordBuffer.Lock(0, 0, d3d9.LOCK_DISCARD)
	check(err)
	texMem.SetFloat32s(0, g.textureCoords)
	check(g.textureCoordBuffer.Unlock())

	check(g.device.SetVertexShader(g.textureVS))
	check(g.device.SetPixelShader(g.texturePS))
	check(g.device.SetVertexDeclaration(g.vertexDecl))
	check(g.device.SetStreamSource(0, g.vertexBuffer, 0, 2*4))
	check(g.device.SetStreamSource(1, g.textureCoordBuffer, 0, 2*4))
	check(g.device.SetTexture(0, g.textureAtlas))
	mvp := d3dmath.Ortho(
		0,
		float32(g.camera.position.W),
		float32(g.camera.position.H),
		0,
		-1,
		1,
	).Transposed()
	check(g.device.SetVertexShaderConstantF(0, mvp[:]))

	check(g.device.BeginScene())
	check(g.device.DrawPrimitive(d3d9.PT_TRIANGLELIST, 0, uint(len(g.vertices)/3)))
	check(g.device.EndScene())

	// clear graphics data for next frame, keep the backing arrays to reduce GC
	// overhead
	g.vertices = g.vertices[:0]
	g.textureCoords = g.textureCoords[:0]
}
