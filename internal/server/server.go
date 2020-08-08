package server

import (
	"context"
	"crypto/sha1"
	"fmt"
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
	go startHttpServer(s.config)
	if err := s.config.loadConfig(); err != nil {
		log.Fatal("error in config loading: ", err.Error())
	}

	log.Info("starting listening on: " + s.config.Port)
	lis, err := net.Listen("tcp", s.config.Port)
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
	resp.Msg = s.config.HttpPort + "/?v=" + strings.Split(fname, "/")[2]
	return &resp, nil
}

func index_handler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("v")
	fmt.Fprintf(w, "<title>Pogogo | %s</title>", filename)
	fmt.Fprintf(w, "<link rel='icon' type='image/ico' href='images/favicon.ico'>")
	fmt.Fprintf(w, "<br><img src='images/%s' style='display:block;margin-left:auto;margin-right:auto;max-width:50%;'>", filename)
	fmt.Fprintf(w, "<center><br><br>")
	fmt.Fprintf(w, "<a href='https://github.com/Acbn-Nick/pogogo' style='text-decoration:none;color:#3e598c;'>Sharing made easy with Pogogo</a>")
	fmt.Fprintf(w, "</center>")

}

func startHttpServer(c *Configuration) {
	http.HandleFunc("/", index_handler)
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./received"))))
	http.ListenAndServe(c.HttpPort, nil)
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
