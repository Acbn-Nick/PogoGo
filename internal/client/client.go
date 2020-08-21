package client

import (
	"context"
	"image"
	"image/png"
	"io/ioutil"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"time"

	"github.com/Acbn-Nick/pogogo/api"
	"github.com/atotto/clipboard"
	"github.com/getlantern/systray"
	"github.com/kbinani/screenshot"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Client struct {
	config *Configuration
	ctx    context.Context
	done   chan interface{}
	reload chan interface{}
	osh    OsHandler
}

type OsHandler interface {
	KeyListen(chan interface{})
	OpenInBrowser(string)
	CaptureArea()
}

func New(ctx context.Context) (*Client, chan interface{}) {
	c := &Client{done: make(chan interface{}), reload: make(chan interface{})}
	c.config = NewConfiguration()
	o := runtime.GOOS
	if o == "windows" {
		c.osh = CreateWinHandler(c)
	} else if o == "darwin" {
		//create macos osHandler
	} else {
		//create linux osHandler
	}
	return c, c.done
}

func (c *Client) Start() {
	log.Info("starting client")
	go systray.Run(c.onReady, c.onExit)
	if err := c.config.loadConfig(); err != nil {
		log.Fatal("error loading config ", err.Error())
	}
	go c.osh.KeyListen(c.reload)
	<-c.done
}

func (c *Client) captureDisplay(n int) {
	time.Sleep(500 * time.Millisecond) //Sleep to wait for clicked menu option to fade away
	bounds := screenshot.GetDisplayBounds(n)
	c.takeScreenshot(bounds)
}

func (c *Client) takeScreenshot(bounds image.Rectangle) {
	img, err := screenshot.CaptureRect(bounds)
	log.Info("capturing: ", bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Max.Y)
	if err != nil {
		log.Info("failed to capture screen ", err.Error())
		return
	}
	t := time.Now()
	fname := t.Format("2006-01-02-15_04_05") + ".png"
	file, _ := os.Create(fname)
	defer file.Close()
	if err != nil {
		log.Info("failed to save screenshot to file ", err.Error())
		return
	}
	if err := png.Encode(file, img); err != nil {
		log.Info("failed to encode screenshot ", err.Error())
		return
	}
	if err := file.Sync(); err != nil {
		log.Info("failed to sync with filesystem ", err.Error())
		return
	}
	if err := c.upload(fname); err != nil {
		log.Info("problem uploading file ", err.Error())
		return
	}
}

func (c *Client) upload(fname string) error {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(c.config.Destination, grpc.WithInsecure())
	if err != nil {
		log.Info("failed to connect to server ", err.Error())
		return err
	}
	defer conn.Close()

	client := api.NewPogogoClient(conn)

	img, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Info("failed to read file ", err.Error())
		return err
	}
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(time.Minute))
	response, err := client.Upload(ctx, &api.UploadRequest{Password: c.config.Password, Image: img})
	if err != nil {
		log.Info("error in request ", err.Error())
		return err
	}
	log.Info("client got response: ", response.Msg)
	if c.config.OpenInBrowser == 1 {
		c.osh.OpenInBrowser("http://" + response.Msg)
	}
	if c.config.CopyToClipboard == 1 {
		clipboard.WriteAll("http://" + response.Msg)
	}
	return nil
}

func (c *Client) onReady() {
	ico, err := ioutil.ReadFile("../server/assets/favicon.ico")
	if err != nil {
		log.Fatal("error loading systray icon ", err.Error())
	}
	time.Sleep(500 * time.Millisecond) // Add 500ms delay to fix issue with systray.AddMenuItem() in goroutines on Windows.
	systray.SetIcon(ico)
	systray.SetTitle("Pogogo")
	systray.SetTooltip("Pogogo Screen Capture")
	cases, numAdded := c.createScreenChans()
	systray.AddSeparator()
	area := systray.AddMenuItem("Snip area", "Snip area")
	systray.AddSeparator()
	browser := systray.AddMenuItem("Open in browser", "Open in browser")
	copy := systray.AddMenuItem("Copy to clipboard", "Copy to clipboard")
	systray.AddSeparator()
	reload := systray.AddMenuItem("Reload config", "Reload config")
	quit := systray.AddMenuItem("Quit", "Quit")
	c.setCheck(browser, c.config.OpenInBrowser, 0)
	c.setCheck(copy, c.config.CopyToClipboard, 1)

	cases[numAdded+1] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(area.ClickedCh)}
	cases[numAdded+2] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(browser.ClickedCh)}
	cases[numAdded+3] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(copy.ClickedCh)}
	cases[numAdded+4] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(reload.ClickedCh)}
	cases[numAdded+5] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(quit.ClickedCh)}

	var (
		areaSnip        = len(cases) - 5
		openInBrowser   = len(cases) - 4
		copyToClipboard = len(cases) - 3
		loadConfig      = len(cases) - 2
		quitSys         = len(cases) - 1
	)

	go func() {
		for {
			chosen, _, _ := reflect.Select(cases)
			if chosen == areaSnip {
				c.osh.CaptureArea()
			} else if chosen == openInBrowser {
				c.setCheck(browser, 1-c.config.OpenInBrowser, 0)
			} else if chosen == copyToClipboard {
				c.setCheck(copy, 1-c.config.CopyToClipboard, 1)
			} else if chosen == loadConfig {
				c.config.loadConfig()
				c.setCheck(browser, c.config.OpenInBrowser, 0)
				c.setCheck(copy, c.config.CopyToClipboard, 1)
				c.reload <- nil
				time.Sleep(1 * time.Second)
				go c.osh.KeyListen(c.reload)
			} else if chosen == quitSys {
				systray.Quit()
				return
			} else {
				c.captureDisplay(chosen)
			}
		}
	}()
}

func (c *Client) setCheck(m *systray.MenuItem, v int, i int) {
	if v == 0 {
		m.Uncheck()
	} else if v == 1 {
		m.Check()
	}
	if i == 0 {
		c.config.OpenInBrowser = v
	} else if i == 1 {
		c.config.CopyToClipboard = v
	}
}

func (c *Client) createScreenChans() ([]reflect.SelectCase, int) {
	var chans []chan struct{}
	for i := 0; i < screenshot.NumActiveDisplays(); i++ {
		mi := systray.AddMenuItem("Capture monitor "+strconv.Itoa(i+1), "Capture monitor "+strconv.Itoa(i+1))
		chans = append(chans, mi.ClickedCh)
	}

	cases := make([]reflect.SelectCase, len(chans)+5)
	numAdded := 0
	for i, ch := range chans {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
		numAdded = i
	}
	return cases, numAdded
}

func (c *Client) onExit() {
	c.done <- nil
	return
}
