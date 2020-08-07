package server

import (
	"context"
	"crypto/sha1"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	api "github.com/Acbn-Nick/pogogo/api"
)

type Server struct {
	HttpServ *http.Server
	config   *Configuration
	files    []string
}

func New() *Server {
	server := Server{config: NewConfiguration()}
	return &server
}

func (s *Server) Start() {
	log.Info("Starting server")

	if err := s.config.loadConfig(); err != nil {
		log.Fatal("Error in config loading: ", err.Error())
	}

	log.Info("Starting listening on: 127.0.0.1" + s.config.Port)
	lis, err := net.Listen("tcp", "127.0.0.1"+s.config.Port)
	if err != nil {
		log.Fatal("Failed to start listening: ", err.Error())
	}

	grpcServer := grpc.NewServer()
	api.RegisterPogogoServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to start gRPC server: ", err.Error())
	}
}

func (s *Server) Upload(ctx context.Context, req *api.UploadRequest) (*api.UploadResponse, error) {
	resp := api.UploadResponse{
		Msg: "",
	}
	log.Info("Received request")
	h := sha1.New()
	if _, err := h.Write([]byte(req.Password)); err != nil {
		log.Info("Failed password hashing on request ", err.Error())
		return &api.UploadResponse{}, err
	}

	hs := string(h.Sum(nil))
	if hs != s.config.Password {
		log.Info("Upload attempted with incorrect password")
		return &api.UploadResponse{}, fmt.Errorf("Incorrect password")
	}

	fname := "./received/" + strconv.FormatInt(time.Now().UTC().UnixNano()/100, 10) + ".png"
	f, err := os.Create(fname)
	if err != nil {
		log.Info("Failed upload ", err.Error())
		return &api.UploadResponse{}, fmt.Errorf("Image upload failed")
	}
	defer f.Close()
	_, err = f.Write(req.Image)
	if err != nil {
		log.Info("Failed upload ", err.Error())
		return &api.UploadResponse{}, fmt.Errorf("Image upload failed")
	}
	err = f.Sync()
	if err != nil {
		log.Info("Failed upload ", err.Error())
		return &api.UploadResponse{}, fmt.Errorf("Image upload failure")
	}
	log.Info("Created file: " + fname)
	resp.Msg = fname
	return &resp, nil
}

func (s *Server) authenticate(pw string) bool {

	return true
}
