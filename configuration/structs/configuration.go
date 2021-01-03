package structs

// Config represents the NSFW service configuration.
type Config struct {
	Cloudmersive *CloudmersiveConfiguration
	Tensorflow   *TensorflowConfiguration
}
