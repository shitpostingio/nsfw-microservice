package structs

// CloudmersiveConfiguration represents the Cloudmersive configuration.
type CloudmersiveConfiguration struct {
	APIKey            string
	APIEndpoint       string
	ExplicitThreshold int
	RacyThreshold     int
}
