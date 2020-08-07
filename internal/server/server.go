package server

import (
	"context"
	"crypto/sha1"
	"fmt"
	"net"
	"net/http"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	api "github.com/Acbn-Nick/pogogo/api"
)

type Server struct {
	HttpServ *http.Server
	config   *Configuration
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

	lis, err := net.Listen("tcp", s.config.Port)
	if err != nil {
		log.Fatal("Failed to start listening: ", err.Error())
	}

	grpcServer := grpc.NewServer()
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to serve gRPC: ", err.Error())
	}

	api.RegisterPogogoServer(grpcServer, s)

	grpcServer.Serve(lis)
}

func (s *Server) Upload(c context.Context, req *api.UploadRequest) (*api.UploadResponse, error) {
	resp := api.UploadResponse{
		Msg: "",
	}

	h := sha1.New()
	if _, err := h.Write([]byte(req.Password)); err != nil {
		log.Info("Failed password hashing on request ", err.Error())
		return nil, err
	}

	hs := string(h.Sum(nil))
	if hs != s.config.Password {
		log.Info("Upload attempted with incorrect password")
		return &api.UploadResponse{}, fmt.Errorf("Incorrect Password")
	}

	return &resp, nil
}
