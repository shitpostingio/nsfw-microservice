package recognition

import (
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/shitpostingio/analysis-commons/decoder"
	"github.com/shitpostingio/analysis-commons/structs"
	nStructs "github.com/shitpostingio/nsfw-microservice/configuration/structs"
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"image"
	"io"
	"log"
	"math"
)

const (

	//
	tensorflowHentaiIndex = 1
	tensorflowPornIndex   = 3
	tensorflowSexyIndex   = 4

	//
	hentaiIndex = 0
	pornIndex   = 1
	sexyIndex   = 2

)

var (
	labels = [...]string{"cartoon porn/hentai", "sexually explicit content/porn", "suggestive content"}
)

// Tensorflow represents a Tensorflow-based NSFW recognition service.
type Tensorflow struct {
	model            *tf.SavedModel
	thresholds       [3]float64
	overallThreshold float64
}

// NewTensorflowInstance instantiates a Tensorflow-based NSFW recognition service.
func NewTensorflowInstance(cfg *nStructs.TensorflowConfiguration) (*Tensorflow, error) {

	model, err := tf.LoadSavedModel(cfg.KnowledgeBasePath, []string{"nsfw-lite"}, nil)
	if err != nil {
		return &Tensorflow{}, fmt.Errorf("NewTensorflowInstance: error loading saved model: %w\n", err)
	}

	return &Tensorflow{
		model:            model,
		thresholds:       [3]float64{cfg.HentaiThreshold, cfg.PornThreshold, cfg.SexyThreshold},
		overallThreshold: cfg.OverallThreshold,
	}, nil

}

// EvaluateImage perform a NSFW recognition on an image using Tensorflow.
func (t *Tensorflow) EvaluateImage(extension string, reader io.Reader, d decoder.MediaDecoder) *structs.Analysis {

	img, err := d.Decode(extension, reader)
	if err != nil {
		log.Println("Tensorflow.EvaluateImage: unable to decode image:", err)
		return &structs.Analysis{NSFWErrorString: err.Error()}
	}

	values, err := t.performInference(img)
	if err != nil {
		log.Println("Tensorflow.EvaluateImage: unable to perform inference on an image:", err)
		return &structs.Analysis{NSFWErrorString: err.Error()}
	}

	return t.convertResults(values)

}

// EvaluateVideo perform a NSFW recognition on a video using Tensorflow.
func (t *Tensorflow) EvaluateVideo(extension string, reader io.Reader, d decoder.MediaDecoder) *structs.Analysis {

	img, err := d.Decode(extension, reader)
	if err != nil {
		log.Println("Tensorflow.EvaluateVideo: unable to decode video:", err)
		return &structs.Analysis{NSFWErrorString: err.Error()}
	}

	values, err := t.performInference(img)
	if err != nil {
		log.Println("Tensorflow.EvaluateVideo: unable to perform inference on a video:", err)
		return &structs.Analysis{NSFWErrorString: err.Error()}
	}

	return t.convertResults(values)

}

func (t *Tensorflow) performInference(img image.Image) (values [3]float32, err error) {

	img = imaging.Resize(img, 224, 224, imaging.NearestNeighbor)
	bounds := img.Bounds()

	// multidim array as input tensor
	var BCHW [1][224][224][3]float32

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			r, g, b, a := img.At(x, y).RGBA()

			fa := float32(a)
			fr := float32(r) / fa
			fg := float32(g) / fa
			fb := float32(b) / fa

			// height = y and width = x
			BCHW[0][y][x][0] = fr
			BCHW[0][y][x][1] = fg
			BCHW[0][y][x][2] = fb
		}
	}

	tensor, err := tf.NewTensor(BCHW)
	if err != nil {
		return values, fmt.Errorf("could not create tensor: %w", err)
	}

	result, err := t.model.Session.Run(
		map[tf.Output]*tf.Tensor{
			t.model.Graph.Operation("input_1").Output(0): tensor,
		},
		[]tf.Output{
			t.model.Graph.Operation("dense_3/Softmax").Output(0),
		},
		nil,
	)

	if err != nil {
		return values, fmt.Errorf("error calculating nsfw values: %w\n", err)

	}

	r := (result[0].Value().([][]float32))[0]
	values[hentaiIndex] = r[tensorflowHentaiIndex]
	values[pornIndex] = r[tensorflowPornIndex]
	values[sexyIndex] = r[tensorflowSexyIndex]
	return

}

func (t *Tensorflow) convertResults(values [3]float32) *structs.Analysis {

	var max, confidence, sum float64
	var label string
	var NaNValues int

	for i := 0; i < 3; i++ {

		currentValue := float64(values[i])
		if math.IsNaN(currentValue) {
			NaNValues++
			continue
		}

		difference := currentValue - t.thresholds[i]
		sum += currentValue
		if difference > max {
			max = difference
			confidence = currentValue
			label = labels[i]
		}

	}

	var response structs.Analysis
	switch {
	case NaNValues == 3:
		response.NSFWErrorString = "all values were NaN"
	case max > 0:
		response.NSFW.IsNSFW = true
		response.NSFW.Confidence = confidence * 100
		response.NSFW.Label = label
	case sum > t.overallThreshold:
		response.NSFW.IsNSFW = true
		response.NSFW.Confidence = sum * 100
		response.NSFW.Label = "potentially unsafe content"
	}

	return &response

}
