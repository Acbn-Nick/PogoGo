package client

import (
	"context"
	"image/png"
	"io/ioutil"
	"os"
	"os/exec"
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

func (c *Client) takeScreenshot() {

	bounds := screenshot.GetDisplayBounds(0)

	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		log.Info("failed to capture screen ", err.Error())
		return
	}
	t := time.Now()
	fname := t.Format("2006-01-02-15,04,05") + ".png"
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
	exec.Command("rundll32", "url.dll,FileProtocolHandler", "http://"+response.Msg).Start()
	return nil
}

func (c *Client) onReady() {
	ico, err := ioutil.ReadFile("../server/assets/favicon.ico")
	if err != nil {
		log.Fatal("error loading systray icon ", err.Error())
	}
	systray.SetIcon(ico)
	systray.SetTitle("Pogogo")
	systray.SetTooltip("Pogogo Screen Capture")
	snip := systray.AddMenuItem("Take screenshot", "Take screenshot")
	reload := systray.AddMenuItem("Reload config", "Reload config")
	quit := systray.AddMenuItem("Quit", "Quit")
	go func() {
		for {
			select {
			case <-snip.ClickedCh:
				c.takeScreenshot()
			case <-reload.ClickedCh:
				c.config.loadConfig()
			case <-quit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func (c *Client) onExit() {
	c.done <- nil
	return
}
