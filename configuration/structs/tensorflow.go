package structs

// TensorflowConfiguration represents the Tensorflow NSFW configuration.
type TensorflowConfiguration struct {
	KnowledgeBasePath string
	HentaiThreshold   float64
	PornThreshold     float64
	SexyThreshold     float64
	OverallThreshold  float64
}
