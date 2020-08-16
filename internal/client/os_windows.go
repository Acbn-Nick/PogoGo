package client

import (
	"image"
	"os/exec"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/Acbn-Nick/pogogo/internal/client/keycode"
	hook "github.com/robotn/gohook"
	log "github.com/sirupsen/logrus"
)

var (
	user32                  = windows.NewLazyDLL("user32.dll")
	procGetAsyncKeyState    = user32.NewProc("GetAsyncKeyState")
	procGetForegroundWindow = user32.NewProc("GetForegroundWindow")
	procGetWindowRect       = user32.NewProc("GetWindowRect")
	procGetCursorPos        = user32.NewProc("GetCursorPos")
	VK_LBUTTON              = 0x01
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

func (o *OsWin) CaptureArea() {
	var (
		lpPointTL = &Point{}
		lpPointBR = &Point{}
	)
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
	held, _, _ := procGetAsyncKeyState.Call(uintptr(VK_LBUTTON))
	for {
		if held == 0 {
			procGetCursorPos.Call(uintptr(unsafe.Pointer(lpPointBR)))
			break
		}
		held, _, _ = procGetAsyncKeyState.Call(uintptr(VK_LBUTTON))
	}
	bounds := image.Rect(int(lpPointTL.x), int(lpPointTL.y), int(lpPointBR.x), int(lpPointBR.y))
	o.c.takeScreenshot(bounds)
}

func (o *OsWin) captureActiveWindow() {
	hwnd, _, _ := procGetForegroundWindow.Call()
	lpRect := &Rect{}
	procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(lpRect)))
	bounds := image.Rect(int(lpRect.left), int(lpRect.top), int(lpRect.right), int(lpRect.bottom))
	o.c.takeScreenshot(bounds)
}
