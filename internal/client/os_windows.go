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

func (o *OsWin) KeyListen() {
	for {
		asynch1, _, _ := procGetAsyncKeyState.Call(uintptr(VK_LCONTROL))
		asynch2, _, _ := procGetAsyncKeyState.Call(uintptr(VK_1))
		asynch3, _, _ := procGetAsyncKeyState.Call(uintptr(VK_LSHIFT))
		if (int)(asynch1) != 0 && (int)(asynch2) != 0 && (int)(asynch3) != 0 {
			log.Info("keys pressed")
			hwnd, _, _ := procGetForegroundWindow.Call()
			lpRect := &Rect{}
			procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(lpRect)))
			bounds := image.Rect(int(lpRect.left), int(lpRect.top), int(lpRect.right), int(lpRect.bottom))
			o.c.takeScreenshot(bounds)
			time.Sleep(1 * time.Second)
		}
		time.Sleep(15 * time.Millisecond)
	}
	return
}

func (o *OsWin) OpenInBrowser(s string) {
	exec.Command("rundll32", "url.dll,FileProtocolHandler", s).Start()
}
