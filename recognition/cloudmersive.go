package recognition

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/shitpostingio/analysis-commons/decoder"
	"github.com/shitpostingio/analysis-commons/encoder"
	"github.com/shitpostingio/analysis-commons/structs"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/rs/xid"
)

const (
	cloudmersiveSuggestiveName = "RacyContent"
)

// Cloudmersive is a struct containing all the data needed to perform requests to Cloudmersive.com.
type Cloudmersive struct {
	CloudmersiveAPIEndpoint string
	CloudmersiveAPIKey      string
	NsfwThreshold           int
	SuggestiveThreshold     int
}

type cloudmersiveNSFWResult struct {
	Successful            bool    `json:"Successful"`
	Score                 float64 `json:"Score"`
	ClassificationOutcome string  `json:"ClassificationOutcome"`
}

// NewCloudmersiveClient starts Cloudmersive and sets some default nsfw and suggestive parameters.
func NewCloudmersiveClient(endpoint string, apikey string, nsfw int, suggestive int) *Cloudmersive {

	return &Cloudmersive{
		CloudmersiveAPIEndpoint: endpoint,
		CloudmersiveAPIKey:      apikey,
		NsfwThreshold:           nsfw,
		SuggestiveThreshold:     suggestive,
	}

}

// EvaluateImage returns the NSFW values of a photo using Cloudmersive's services.
func (c *Cloudmersive) EvaluateImage(extension string, file io.Reader, decoder decoder.MediaDecoder) *structs.Analysis {

	var name string
	if extension == "webp" {
		img, err := decoder.Decode(extension, file)
		if err != nil {
			log.Println("Cloudmersive.EvaluateImage: unable to decode webp image:", err)
			return &structs.Analysis{NSFWErrorString: err.Error()}
		}

		name = fmt.Sprintf("%s.%s", xid.New(), "png")
		filename := fmt.Sprintf("/tmp/%s", name)
		err = encoder.SaveImageAsPNG(filename, img)
		if err != nil {
			log.Println("Cloudmersive.EvaluateImage: unable to save converted webp on disk ", err)
			return &structs.Analysis{NSFWErrorString: err.Error()}
		}

		file, err = os.Open(filename)
		if err != nil {
			log.Println("Cloudmersive.EvaluateImage: unable to open file", err)
			return &structs.Analysis{NSFWErrorString: err.Error()}
		}

		defer func() {
			err := os.Remove(filename)
			if err != nil {
				log.Println("Cloudmersive.EvaluateImage: unable to remove file", err)
			}
		}()

	} else {
		name = fmt.Sprintf("%s.%s", xid.New(), extension)
	}

	cloudmersiveResult, err := c.performRequest(file, name)
	if err != nil || !cloudmersiveResult.Successful {
		return &structs.Analysis{NSFWErrorString: fmt.Sprintf("Cloudmersive.EvaluateImage: unable to get NSFW result: %s", err)}
	}

	return &structs.Analysis{NSFW: c.convertCloudmersiveResult(cloudmersiveResult)}

}

// EvaluateVideo returns the NSFW values of a video using Cloudmersive's services.
func (c *Cloudmersive) EvaluateVideo(extension string, file io.Reader, d decoder.MediaDecoder) *structs.Analysis {

	img, err := d.Decode(extension, file)
	if err != nil {
		log.Println("Cloudmersive.EvaluateVideo: unable to decode video:", err)
		return &structs.Analysis{NSFWErrorString: err.Error()}
	}

	name := fmt.Sprintf("%s.%s", xid.New(), "png")
	filename := fmt.Sprintf("/tmp/%s", name)
	err = encoder.SaveImageAsPNG(filename, img)
	if err != nil {
		log.Println("Cloudmersive.EvaluateVideo: unable to save frame on disk", err)
		return &structs.Analysis{NSFWErrorString: err.Error()}
	}

	file, err = os.Open(filename)
	if err != nil {
		log.Println("Cloudmersive.EvaluateVideo: unable to open frame", err)
		return &structs.Analysis{NSFWErrorString: err.Error()}
	}

	defer func() {
		err := os.Remove(filename)
		if err != nil {
			log.Println("Cloudmersive.EvaluateVideo: unable to remove frame", err)
		}
	}()

	return c.EvaluateImage("png", file, &decoder.ImageDecoder{})

}

func (c *Cloudmersive) performRequest(data io.Reader, filename string) (result cloudmersiveNSFWResult, err error) {

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("imageFile", filename)
	if err != nil {
		return
	}

	_, err = io.Copy(part, data)
	if err != nil {
		return
	}

	err = writer.Close()
	if err != nil {
		return
	}

	request, err := http.NewRequest(http.MethodPost, c.CloudmersiveAPIEndpoint, body)
	if err != nil {
		return
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Apikey", c.CloudmersiveAPIKey)

	client := http.Client{Timeout: time.Second * 30}
	response, err := client.Do(request)
	if err != nil {
		return
	}

	defer func() {
		err = response.Body.Close()
		if err != nil {
			log.Println("Cloudmersive.performRequest: unable to close response body")
		}
	}()

	bodyResult, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(bodyResult, &result)
	return

}

func (c *Cloudmersive) convertCloudmersiveResult(cloudmersiveResult cloudmersiveNSFWResult) (r structs.NSFWResponse) {

	r.Confidence = cloudmersiveResult.Score * 100
	if cloudmersiveResult.ClassificationOutcome == cloudmersiveSuggestiveName {
		r.IsNSFW = int(r.Confidence)+1 > c.SuggestiveThreshold
	} else {
		r.IsNSFW = int(r.Confidence)+1 > c.NsfwThreshold
	}

	r.Label = cloudmersiveResult.ClassificationOutcome
	return
}
