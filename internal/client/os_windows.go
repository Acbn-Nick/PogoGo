package client

import (
	"image"
	"image/png"
	"os"
	"os/exec"
	"runtime"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/Acbn-Nick/pogogo/internal/client/keycode"
	"github.com/kbinani/screenshot"
	"github.com/lxn/win"
	hook "github.com/robotn/gohook"
	log "github.com/sirupsen/logrus"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

var (
	user32                  = windows.NewLazyDLL("user32.dll")
	procGetAsyncKeyState    = user32.NewProc("GetAsyncKeyState")
	procGetForegroundWindow = user32.NewProc("GetForegroundWindow")
	procGetWindowRect       = user32.NewProc("GetWindowRect")
	procGetCursorPos        = user32.NewProc("GetCursorPos")
	procGetSystemMetrics    = user32.NewProc("GetSystemMetrics")
	procBringWindowToTop    = user32.NewProc("BringWindowToTop")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	VK_LBUTTON              = 0x01
	VK_ESCAPE               = 0x1B
)

type Rect struct {
	left   int32
	top    int32
	right  int32
	bottom int32
}

type Point struct {
	x int32
	y int32
}

type OsWin struct {
	c *Client
}

func CreateWinHandler(c *Client) OsHandler {
	return &OsWin{c: c}
}

func (o *OsWin) startHooks(capActive []string, capSnip []string, dd chan interface{}) {
	log.Info("starting hooks")
	hook.Register(hook.KeyDown, capActive, func(e hook.Event) {
		o.captureActiveWindow()
	})
	hook.Register(hook.KeyDown, capSnip, func(e hook.Event) {
		o.CaptureArea()
	})
	s := hook.Start()
	<-hook.Process(s)
}

func (o *OsWin) KeyListen(done chan interface{}) {
	capActive, err := keycode.ParseShortcut(o.c.config.CaptureActive)
	if err != nil {
		log.Infof("failed to parse capture active shortcut ", err.Error())
		return
	}
	capSnip, err := keycode.ParseShortcut(o.c.config.CaptureSnip)
	if err != nil {
		log.Infof("failed to parse capture snip shortcut ", err.Error())
		return
	}
	dd := make(chan interface{})
	go o.startHooks(capActive, capSnip, dd)

	<-done
	log.Info("killing hooks")
	hook.End()
	return
}

func (o *OsWin) OpenInBrowser(s string) {
	exec.Command("rundll32", "url.dll,FileProtocolHandler", s).Start()
}

type DWM_BLURBEHIND struct {
	DwFlags                uint32
	fEnable                int32
	hRgnBlur               uint64
	fTransitionOnMaximized int32
}

func (o *OsWin) CaptureArea() {
	var (
		lpPointTL = &Point{}
		lpPointBR = &Point{}
		r         = sdl.Rect{X: 0, Y: 0, W: 0, H: 0}
	)
	window, renderer, tex, err := o.createSdlWindow()

	defer window.Destroy()
	defer tex.Destroy()
	defer renderer.Destroy()
	defer sdl.Quit()

	renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)

	wmi, err := window.GetWMInfo()
	if err != nil {
		log.Info("failed to get window manager info for window ", err.Error())
	}
	hwnd := wmi.GetWindowsInfo().Window
	procBringWindowToTop.Call(uintptr(hwnd))
	procSetForegroundWindow.Call(uintptr(hwnd))

	renderer.SetDrawColor(42, 82, 190, 255)

	cursor := sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_CROSSHAIR)
	defer sdl.FreeCursor(cursor)
	sdl.SetCursor(cursor)
	procGetAsyncKeyState.Call(uintptr(VK_LBUTTON))

	click, _, _ := procGetAsyncKeyState.Call(uintptr(VK_LBUTTON))
	click &= 0x100
	for {
		if esc, _, _ := procGetAsyncKeyState.Call(uintptr(VK_ESCAPE)); esc != 0 {
			return
		}
		if click != 0 {
			procGetCursorPos.Call(uintptr(unsafe.Pointer(lpPointTL)))
			r.X = lpPointTL.x
			r.Y = lpPointTL.y
			break
		}
		time.Sleep(15 * time.Millisecond)
		click, _, _ = procGetAsyncKeyState.Call(uintptr(VK_LBUTTON))
	}

	if err != nil {
		log.Info(err.Error)
	}
	held, _, _ := procGetAsyncKeyState.Call(uintptr(VK_LBUTTON))
	for {
		renderer.Clear()
		if esc, _, _ := procGetAsyncKeyState.Call(uintptr(VK_ESCAPE)); esc != 0 {
			return
		}
		if held == 0 {
			procGetCursorPos.Call(uintptr(unsafe.Pointer(lpPointBR)))
			renderer.Clear()
			renderer.Copy(tex, nil, nil)
			renderer.Present()
			break
		}
		procGetCursorPos.Call(uintptr(unsafe.Pointer(lpPointBR)))
		r.W = lpPointBR.x - lpPointTL.x
		r.H = lpPointBR.y - lpPointTL.y
		renderer.Copy(tex, nil, nil)
		renderer.SetDrawColor(42, 82, 190, 255)
		renderer.DrawRect(&r)
		renderer.SetDrawColor(42, 82, 190, 30)
		renderer.FillRect(&r)
		renderer.Present()
		time.Sleep(10 * time.Millisecond)
		held, _, _ = procGetAsyncKeyState.Call(uintptr(VK_LBUTTON))
	}
	bounds := image.Rect(int(lpPointTL.x), int(lpPointTL.y), int(lpPointBR.x), int(lpPointBR.y))
	o.c.takeScreenshot(bounds)
	return
}

func (o *OsWin) captureActiveWindow() {
	hwnd, _, _ := procGetForegroundWindow.Call()
	lpRect := &Rect{}
	procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(lpRect)))
	bounds := image.Rect(int(lpRect.left), int(lpRect.top), int(lpRect.right), int(lpRect.bottom))
	o.c.takeScreenshot(bounds)
}

func (o *OsWin) createSdlWindow() (*sdl.Window, *sdl.Renderer, *sdl.Texture, error) {
	runtime.LockOSThread()
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		log.Info("failed to initialize sdl ", err.Error())
		return nil, nil, nil, err
	}

	cx := win.GetSystemMetrics(win.SM_CXVIRTUALSCREEN)
	cy := win.GetSystemMetrics(win.SM_CYVIRTUALSCREEN)

	window, err := sdl.CreateWindow("PogogoWin", 0, 0, cx, cy, sdl.WINDOW_BORDERLESS|sdl.WINDOW_MOUSE_CAPTURE|sdl.WINDOW_ALWAYS_ON_TOP)
	if err != nil {
		log.Info("failed to create sdl window ", err.Error())
		return nil, nil, nil, err
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_SOFTWARE)
	if err != nil {
		log.Info("failed to create renderer ", err.Error())
	}

	imga, err := screenshot.CaptureRect(image.Rect(0, 0, int(cx), int(cy)))
	file, err := os.Create("t1.png")
	if err != nil {
		log.Info("error creating file ", err.Error())
	}
	if err != nil {
		log.Info("failed to save screenshot to file ", err.Error())
	}
	if err := png.Encode(file, imga); err != nil {
		log.Info("failed to encode screenshot ", err.Error())
	}
	if err := file.Sync(); err != nil {
		log.Info("failed to sync with filesystem ", err.Error())
	}
	if err := file.Close(); err != nil {
		log.Info("failed to close file handle ", err.Error())
	}
	sImg, err := img.Load("t1.png")
	defer sImg.Free()
	tex, err := renderer.CreateTextureFromSurface(sImg)
	if err != nil {
		log.Info("error in creating texture ", err.Error())
	}
	renderer.Clear()
	renderer.Copy(tex, nil, nil)
	renderer.Present()
	os.Remove("t1.png")
	return window, renderer, tex, err
}
