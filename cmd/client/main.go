package main

import (
	"context"
	"io/ioutil"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	api "github.com/Acbn-Nick/pogogo/api"
)

func main() {
	//Below is placeholder code until client is implemented,
	// this is just to test communication.

	log.Info("starting client")
	var conn *grpc.ClientConn
	conn, err := grpc.Dial("127.0.0.1:9001", grpc.WithInsecure())
	if err != nil {
		log.Fatal("failed to connect to server ", err.Error())
	}
	defer conn.Close()

	c := api.NewPogogoClient(conn)

	img, err := ioutil.ReadFile("./gnu.png")
	if err != nil {
		log.Info("failed to read file ", err.Error())
	}

	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(time.Minute))
	response, err := c.Upload(ctx, &api.UploadRequest{Password: "pogogo", Image: img})
	if err != nil {
		log.Fatal("error in request ", err.Error())
	}
	log.Info("client got response: ", response.Msg)
	exec.Command("rundll32", "url.dll,FileProtocolHandler", "http://"+response.Msg).Start()
}
