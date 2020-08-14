package client

import (
	"image"
	"os/exec"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	log "github.com/sirupsen/logrus"
)

var (
	user32                  = windows.NewLazyDLL("user32.dll")
	procGetAsyncKeyState    = user32.NewProc("GetAsyncKeyState")
	procGetForegroundWindow = user32.NewProc("GetForegroundWindow")
	procGetWindowRect       = user32.NewProc("GetWindowRect")
	VK_LCONTROL             = 0xA2
	VK_LSHIFT               = 0xA0
	VK_1                    = 0x31
	VK_Q                    = 0x51
)

type Rect struct {
	left   int32
	top    int32
	right  int32
	bottom int32
}

type OsWin struct {
	c *Client
}

func CreateWinHandler(c *Client) OsHandler {
	return &OsWin{c: c}
}

func (o *OsWin) KeyListen(done chan interface{}) {
	capActive, err := parseShortcut(o.c.config.CaptureActive)
	if err != nil {
		log.Infof("failed to parse capture active shortcut ", err.Error())
		return
	}
	capSnip, err := parseShortcut(o.c.config.CaptureSnip)
	if err != nil {
		log.Infof("failed to parse capture snip shortcut ", err.Error())
		return
	}
	for {
		select {
		case <-done:
			return
		default:
			activePressed := true
			for _, v := range capActive {
				async, _, _ := procGetAsyncKeyState.Call(uintptr(uintptr(v)))
				activePressed = activePressed && (async != 0)
			}

			capPressed := true
			for _, v := range capSnip {
				async, _, _ := procGetAsyncKeyState.Call(uintptr(uintptr(v)))
				capPressed = capPressed && (async != 0)
			}

			if activePressed {
				o.captureActiveWindow()
				time.Sleep(1 * time.Second)
			} else if capPressed {
				o.CaptureArea()
				time.Sleep(1 * time.Second)
			}

			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (o *OsWin) OpenInBrowser(s string) {
	exec.Command("rundll32", "url.dll,FileProtocolHandler", s).Start()
}

func (o *OsWin) CaptureArea() {

}

func (o *OsWin) captureActiveWindow() {
	hwnd, _, _ := procGetForegroundWindow.Call()
	lpRect := &Rect{}
	procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(lpRect)))
	bounds := image.Rect(int(lpRect.left), int(lpRect.top), int(lpRect.right), int(lpRect.bottom))
	o.c.takeScreenshot(bounds)
}
