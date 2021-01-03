package client

import (
	"bytes"
	"encoding/json"
	"github.com/shitpostingio/analysis-commons/structs"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"
)

// PerformRequest performs a request to the NSFW service.
func PerformRequest(file io.Reader, fileName, endpoint string) (data structs.NSFWResponse, errorString string) {

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	//
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		errorString = err.Error()
		return
	}

	_, err = io.Copy(part, file)
	if err != nil {
		errorString = err.Error()
		return
	}

	err = writer.Close()
	if err != nil {
		errorString = err.Error()
		return
	}

	request, err := http.NewRequest(http.MethodPost, endpoint, body)
	if err != nil {
		errorString = err.Error()
		return
	}

	// We want to send data with the multipart/form-data Content-Type
	request.Header.Add("Content-Type", writer.FormDataContentType())
	client := http.Client{Timeout: time.Second * 30}
	response, err := client.Do(request)
	if err != nil {
		errorString = err.Error()
		return
	}

	defer func() {
		err = response.Body.Close()
		if err != nil {
			log.Println("NSFWClient.PerformRequest: unable to close response body")
		}
	}()

	bodyResult, err := ioutil.ReadAll(response.Body)
	log.Debugln("NSFWClient request result: ", string(bodyResult))
	if err != nil {
		errorString = err.Error()
		return
	}

	var nr structs.Analysis
	err = json.Unmarshal(bodyResult, &nr)
	if err != nil {
		errorString = err.Error()
		return
	}

	return nr.NSFW, nr.NSFWErrorString

}
