package client

import (
	"context"
	"image/png"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"time"

	"github.com/Acbn-Nick/pogogo/api"
	"github.com/getlantern/systray"
	"github.com/kbinani/screenshot"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Client struct {
	config *Configuration
	ctx    context.Context
	done   chan interface{}
}

func New(ctx context.Context) (*Client, chan interface{}) {
	c := &Client{done: make(chan interface{})}
	c.config = NewConfiguration()
	return c, c.done
}

func (c *Client) Start() {
	log.Info("starting client")
	go systray.Run(c.onReady, c.onExit)
	if err := c.config.loadConfig(); err != nil {
		log.Fatal("error loading config ", err.Error())
	}
	<-c.done
}

func (c *Client) takeScreenshot(n int) {
	time.Sleep(500 * time.Millisecond) //Sleep to wait for clicked menu option to fade away
	bounds := screenshot.GetDisplayBounds(n)
	img, err := screenshot.CaptureRect(bounds)
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
		exec.Command("rundll32", "url.dll,FileProtocolHandler", "http://"+response.Msg).Start()
	}
	if c.config.CopyToClipboard == 1 {
		//Unimplemented
	}
	return nil
}

func (c *Client) onReady() {
	ico, err := ioutil.ReadFile("../server/assets/favicon.ico")
	if err != nil {
		log.Fatal("error loading systray icon ", err.Error())
	}
	time.Sleep(500 * time.Millisecond) // Add 500ms delay to fix issue with systray.AddMenuItem() in go routines.
	systray.SetIcon(ico)
	systray.SetTitle("Pogogo")
	systray.SetTooltip("Pogogo Screen Capture")

	var chans []chan struct{}
	for i := 0; i < screenshot.NumActiveDisplays(); i++ {
		mi := systray.AddMenuItem("Capture monitor "+strconv.Itoa(i+1), "Capture monitor "+strconv.Itoa(i+1))
		chans = append(chans, mi.ClickedCh)
	}

	cases := make([]reflect.SelectCase, len(chans)+4)
	numAdded := 0
	for i, ch := range chans {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
		numAdded = i
	}
	systray.AddSeparator()
	browser := systray.AddMenuItem("Open in browser", "Open in browser")
	if c.config.OpenInBrowser == 1 {
		browser.Check()
	}
	cases[numAdded+1] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(browser.ClickedCh)}
	copy := systray.AddMenuItem("Copy to clipboard", "Copy to clipboard")
	if c.config.CopyToClipboard == 1 {
		copy.Check()
	}
	cases[numAdded+2] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(copy.ClickedCh)}
	systray.AddSeparator()
	reload := systray.AddMenuItem("Reload config", "Reload config")
	cases[numAdded+3] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(reload.ClickedCh)}
	quit := systray.AddMenuItem("Quit", "Quit")
	cases[numAdded+4] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(quit.ClickedCh)}
	var (
		openInBrowser   = len(cases) - 4
		copyToClipboard = len(cases) - 3
		loadConfig      = len(cases) - 2
		quitSys         = len(cases) - 1
	)
	go func() {
		for {
			chosen, _, _ := reflect.Select(cases)
			if chosen == openInBrowser {
				if browser.Checked() {
					browser.Uncheck()
					c.config.OpenInBrowser = 0
				} else {
					browser.Check()
					c.config.OpenInBrowser = 1
				}
			} else if chosen == copyToClipboard {
				if copy.Checked() {
					copy.Uncheck()
					c.config.CopyToClipboard = 0
				} else {
					copy.Check()
					c.config.CopyToClipboard = 1
				}
			} else if chosen == loadConfig {
				c.config.loadConfig()
				if c.config.OpenInBrowser == 1 {
					browser.Check()
				} else {
					browser.Uncheck()
				}
				if c.config.CopyToClipboard == 1 {
					copy.Check()
				} else {
					copy.Uncheck()
				}
			} else if chosen == quitSys {
				systray.Quit()
				return
			} else {
				c.takeScreenshot(chosen)
			}
		}
	}()
}

func (c *Client) onExit() {
	c.done <- nil
	return
}
