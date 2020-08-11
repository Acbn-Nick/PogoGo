package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	server "github.com/Acbn-Nick/pogogo/internal/client"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	c, done := server.New(ctx)

	go c.Start()

	//<-sigs
	<-done
	log.Info("killing client")
	cancel()
}

/*var conn *grpc.ClientConn
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
*/
