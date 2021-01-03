package recognition

import (
	"github.com/shitpostingio/analysis-commons/decoder"
	"github.com/shitpostingio/analysis-commons/structs"
	"io"
)

// Provider represents a generic NSFW recognition provider.
type Provider interface {
	EvaluateImage(extension string, reader io.Reader, d decoder.MediaDecoder) *structs.Analysis
	EvaluateVideo(extension string, reader io.Reader, d decoder.MediaDecoder) *structs.Analysis
}
