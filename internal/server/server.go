package server

import (
	"context"
	"crypto/sha1"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	log.Info("starting grpc and http server")
	go startHttpServer()
	if err := s.config.loadConfig(); err != nil {
		log.Fatal("error in config loading: ", err.Error())
	}

	log.Info("starting listening on: 127.0.0.1" + s.config.Port)
	lis, err := net.Listen("tcp", "127.0.0.1"+s.config.Port)
	if err != nil {
		log.Fatal("failed to start listening: ", err.Error())
	}

	grpcServer := grpc.NewServer()
	api.RegisterPogogoServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("failed to start gRPC server: ", err.Error())
	}
}

func (s *Server) Upload(ctx context.Context, req *api.UploadRequest) (*api.UploadResponse, error) {
	resp := api.UploadResponse{
		Msg: "",
	}

	log.Info("received request")

	if !s.authenticate(req.Password) {
		log.Info("upload attempted with incorrect password")
		return &api.UploadResponse{}, fmt.Errorf("incorrect password")
	}

	fname, err := s.addFile(req.Image)
	if err != nil {
		log.Info("failed upload ", err.Error())
		return &api.UploadResponse{}, err
	}

	log.Info("created file: " + fname)
	resp.Msg = fname
	return &resp, nil
}

func startHttpServer() {
	tmpl := template.Must(template.ParseFiles("../../web/page.html"))
	if err := http.ListenAndServe("127.0.0.1:8080", handle(tmpl)); err != nil {
		log.Fatalf("error starting http server: ", err.Error())
	}
}

func handle(t *template.Template) http.Handler {
	hdl := func(rw http.ResponseWriter, r *http.Request) {
		if err := t.Execute(rw, r.URL.Query()); err != nil {
			http.Error(rw, "error executing template "+err.Error(), http.StatusInternalServerError)
		}
	}

	return http.HandlerFunc(hdl)

}

func (s *Server) addFile(img []byte) (string, error) {
	fname := "./received/" + strconv.FormatInt(time.Now().UTC().UnixNano()/100, 10) + ".png"
	f, err := os.Create(fname)
	if err != nil {
		return "", fmt.Errorf("image upload failed")
	}
	defer f.Close()
	_, err = f.Write(img)
	if err != nil {
		return "", fmt.Errorf("image upload failed")
	}
	err = f.Sync()
	if err != nil {
		return "", fmt.Errorf("image upload failed")
	}
	if s.config.Ttl != 0 {
		go s.trackAndCull(fname)
	}
	return fname, nil
}

func (s *Server) authenticate(pw string) bool {
	h := sha1.New()
	if _, err := h.Write([]byte(pw)); err != nil {
		log.Info("failed password hashing on request ", err.Error())
		return false
	}
	hs := string(h.Sum(nil))
	return hs == s.config.Password
}

func (s *Server) trackAndCull(fn string) {
	s.files = append(s.files, fn)
	t := time.Now()
	removed := 0
	for i, si := range s.files {
		name := strings.Split(si, "/")[2]
		name = name[:len(name)-4]
		nano, _ := strconv.Atoi(name)
		created := time.Unix(0, int64(nano*100))
		elapsed := t.Sub(created)
		if elapsed >= s.config.Ttl {
			if err := os.Remove(s.files[i]); err != nil {
				log.Info("file culling failed ", err.Error())
			}
			log.Info("removed file: ", s.files[i])
			removed++
		} else {
			break
		}
	}
	s.files = s.files[removed:]
}
