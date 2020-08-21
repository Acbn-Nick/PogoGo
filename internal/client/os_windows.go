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
	hook "github.com/robotn/gohook"
	log "github.com/sirupsen/logrus"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

var (
	user32                         = windows.NewLazyDLL("user32.dll")
	dwmapi                         = windows.NewLazyDLL("dwmapi.dll")
	procGetAsyncKeyState           = user32.NewProc("GetAsyncKeyState")
	procGetForegroundWindow        = user32.NewProc("GetForegroundWindow")
	procGetWindowRect              = user32.NewProc("GetWindowRect")
	procGetCursorPos               = user32.NewProc("GetCursorPos")
	procGetDesktopWindow           = user32.NewProc("GetDesktopWindow")
	procSetWindowLong              = user32.NewProc("SetWindowLongA")
	procGetWindowLong              = user32.NewProc("GetWindowLongA")
	procSetLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")
	procDwmEnableBlurBehindWindow  = dwmapi.NewProc("DwmEnableBlurBehindWindow")
	VK_LBUTTON                     = 0x01
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
	)
	window, renderer, err := o.createSdlWindow()

	var r = sdl.Rect{X: 300, Y: 300, W: 75, H: 75}
	/*renderer.SetDrawColor(255, 255, 00, 255)
	renderer.DrawRect(&r)*/
	renderer.SetDrawColor(40, 128, 40, 255)
	renderer.FillRect(&r)

	/*hwin := win.GetActiveWindow()

	wl := win.GetWindowLong(hwin, win.GWL_EXSTYLE)
	wl = wl | win.WS_EX_LAYERED
	log.Info("window handle: ", hwin)
	if x := win.SetWindowLong(hwin, win.GWL_EXSTYLE, wl); x == 0 {
		log.Info("failed to set window long")
	}

	procSetLayeredWindowAttributes.Call(uintptr(hwin), 0x00ffff00, 10, 0x00000001)

	bb := DWM_BLURBEHIND{DwFlags: 0x00000005, fEnable: 1, hRgnBlur: 0, fTransitionOnMaximized: 1}

	procDwmEnableBlurBehindWindow.Call(uintptr(hwin), uintptr(unsafe.Pointer(&bb)))
	win.UpdateWindow(hwin)

	win.SetWindowPos(hwin, hwin, 0, 0, 1, 1, win.SWP_NOMOVE|win.SWP_NOSIZE|win.SWP_NOZORDER|win.SWP_FRAMECHANGED)
	window.SetWindowOpacity(0.1)
	window.UpdateSurface()
	renderer.Present()*/

	cursor := sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_CROSSHAIR)
	defer sdl.FreeCursor(cursor)
	sdl.SetCursor(cursor)
	procGetAsyncKeyState.Call(uintptr(VK_LBUTTON))
	time.Sleep(100 * time.Millisecond)
	click, _, _ := procGetAsyncKeyState.Call(uintptr(VK_LBUTTON))
	click &= 0x100
	for {
		if click != 0 {
			procGetCursorPos.Call(uintptr(unsafe.Pointer(lpPointTL)))
			break
		}
		click, _, _ = procGetAsyncKeyState.Call(uintptr(VK_LBUTTON))
	}

	if err != nil {
		log.Info(err.Error)
	}
	held, _, _ := procGetAsyncKeyState.Call(uintptr(VK_LBUTTON))
	for {
		if held == 0 {
			procGetCursorPos.Call(uintptr(unsafe.Pointer(lpPointBR)))
			break
		}
		held, _, _ = procGetAsyncKeyState.Call(uintptr(VK_LBUTTON))
	}
	bounds := image.Rect(int(lpPointTL.x), int(lpPointTL.y), int(lpPointBR.x), int(lpPointBR.y))
	window.Destroy()
	sdl.Quit()
	time.Sleep(200 * time.Millisecond)
	o.c.takeScreenshot(bounds)
}

func (o *OsWin) captureActiveWindow() {
	hwnd, _, _ := procGetForegroundWindow.Call()
	lpRect := &Rect{}
	procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(lpRect)))
	bounds := image.Rect(int(lpRect.left), int(lpRect.top), int(lpRect.right), int(lpRect.bottom))
	o.c.takeScreenshot(bounds)
}

func (o *OsWin) createSdlWindow() (*sdl.Window, *sdl.Renderer, error) {
	runtime.LockOSThread()
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		log.Info("failed to initialize sdl ", err.Error())
		return nil, nil, err
	}
	hwnd, _, _ := procGetDesktopWindow.Call()
	var rect = &Rect{0, 0, 0, 0}
	procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(rect)))

	window, err := sdl.CreateWindow("PogogoWin", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, rect.right-rect.left, rect.bottom-rect.top, sdl.WINDOW_FULLSCREEN|sdl.WINDOW_MOUSE_CAPTURE|sdl.WINDOW_ALWAYS_ON_TOP)
	if err != nil {
		log.Info("failed to create sdl window ", err.Error())
		return nil, nil, err
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_SOFTWARE)
	if err != nil {
		log.Info("failed to create renderer ", err.Error())
	}

	imga, err := screenshot.CaptureDisplay(0)
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
	file.Close()
	sImg, err := img.Load("t1.png")

	tex, err := renderer.CreateTextureFromSurface(sImg)
	if err != nil {
		log.Info("error in creating texture ", err.Error())
	}
	renderer.Clear()
	renderer.Copy(tex, nil, nil)
	renderer.Present()
	os.Remove("t1.png")
	return window, renderer, err
}
