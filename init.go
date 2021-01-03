package main

import (
	"github.com/gorilla/mux"
	"github.com/shitpostingio/nsfw-microservice/configuration"
	"github.com/shitpostingio/nsfw-microservice/recognition"
	"log"
	"os"
	"strconv"
)

func init() {

	setEnvVars()
	cfg, err := configuration.Load(configurationPath)
	if err != nil {
		log.Fatal(err)
	}

	switch nsfwType {
	case "t":

		t, err = recognition.NewTensorflowInstance(cfg.Tensorflow)
		if err != nil {
			log.Fatal(err)
		}

		imageHandler = t.EvaluateImage
		videoHandler = t.EvaluateVideo

	case "c":

		cmCfg := cfg.Cloudmersive
		c = recognition.NewCloudmersiveClient(cmCfg.APIEndpoint, cmCfg.APIKey, cmCfg.ExplicitThreshold, cmCfg.RacyThreshold)

		imageHandler = c.EvaluateImage
		videoHandler = c.EvaluateVideo

	default:

		log.Fatal("NsfwType must be either c or t")

	}

	r = mux.NewRouter()

}

func setEnvVars() {

	add := os.Getenv(bindAddressKey)
	if add != "" {
		bindAddress = add
	}

	cp := os.Getenv(cfgPathKey)
	if cp != "" {
		configurationPath = cp
	}

	nt := os.Getenv(nsfwTypeKey)
	if nt != "" {
		nsfwType = nt
	}

	mis := os.Getenv(imageSizeKey)
	if mis != "" {
		p, err := strconv.ParseInt(mis, 10, 64)
		if p > 0 && err == nil {
			maxImageSize = p
		}
	}

	mvs := os.Getenv(videoSizeKey)
	if mvs != "" {
		p, err := strconv.ParseInt(mvs, 10, 64)
		if p > 0 && err == nil {
			maxVideoSize = p
		}
	}

}
