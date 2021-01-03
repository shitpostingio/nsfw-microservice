package main

import (
	"github.com/gorilla/mux"
	"github.com/shitpostingio/analysis-commons/decoder"
	"github.com/shitpostingio/analysis-commons/handler"
	health_check "github.com/shitpostingio/analysis-commons/health-check"
	"github.com/shitpostingio/nsfw-microservice/recognition"
	"log"
	"net/http"
)

const (
	bindAddressKey = "NSFW_BIND_ADDRESS"
	cfgPathKey     = "NSFW_CFG_PATH"
	nsfwTypeKey    = "NSFW_TYPE"
	imageSizeKey   = "NSFW_MAX_IMAGE_SIZE"
	videoSizeKey   = "NSFW_MAX_VIDEO_SIZE"
)

var (
	bindAddress             = "localhost:10001"
	configurationPath       = "config.toml"
	nsfwType                = "t"
	maxImageSize      int64 = 10 << 20 // 10MB
	maxVideoSize      int64 = 20 << 20 // 20MB

	r            *mux.Router
	c            *recognition.Cloudmersive
	t            *recognition.Tensorflow
	imageHandler handler.Handler
	videoHandler handler.Handler
)

func main() {

	r.HandleFunc("/nsfw/image", handleNSFWImage).Methods("POST")
	r.HandleFunc("/nsfw/video", handleNSFWVideo).Methods("POST")
	r.HandleFunc("/healthy", health_check.ConfirmServiceHealth)
	log.Println("NSFW recognition server powered on!")
	log.Fatal(http.ListenAndServe(bindAddress, r))

}

func handleNSFWImage(w http.ResponseWriter, r *http.Request) {
	handler.Handle(w, r, maxImageSize, &decoder.ImageDecoder{}, imageHandler)
}

func handleNSFWVideo(w http.ResponseWriter, r *http.Request) {
	handler.Handle(w, r, maxVideoSize, &decoder.VideoDecoder{}, videoHandler)
}
