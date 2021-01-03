# NSFW Microservice

REST-powered microservice for analyzing media and returning NSFW scores.
The repository also contains a client to perform requests to the service.

It is important to notice that this service provides no support for authentication or caching. It is also completely stateless, making it ideal to be used in the backend. A possible "frontend" implementation can be found in [Analysis API](https://gitlab.com/shitposting/analysis-api).

## Endpoints

- Image endpoint: `<bind-address>/nsfw/image`
- Video endpoint: `<bind-address>/nsfw/video`
- Health check: `<bind-address>/healthy`

## Returned data

The data returned by the server is in the form:

```go
type Analysis struct {
    Fingerprint            FingerprintResponse
    NSFW                   NSFWResponse
    FingerprintErrorString string
    NSFWErrorString        string
}

```

The client trims off the unnecessary data and returns:

```go
type NSFWResponse struct {
    IsNSFW     bool
    Confidence float64
    Label      string
}
```

## Environment options

- Service bind address and port: `NSFW_BIND_ADDRESS` (defaults to `localhost:10001`).
- Path to configuration file: `NSFW_CFG_PATH` (defaults to `config.toml`).
- Recognition service to use: `NSFW_TYPE`. Currently supported values are `t` (Tensorflow, default) and `c` (Cloudmersive).
- Max size for image files: `NSFW_MAX_IMAGE_SIZE` (defaults to `10 << 20`, 10 MB).
- Max size for video files: `NSFW_MAX_VIDEO_SIZE` (defaults to `20 << 20`, 20 MB).

## Configuration file structure

```toml
[cloudmersive]
  apiendpoint = "https://api.cloudmersive.com/image/nsfw/classify"
  apikey = <your-cloudmersive-api-key>
  explicitthreshold = 70
  racythreshold = 78

[tensorflow]
  knowledgebasepath = <path-to-folder-containing-model>
  hentaiThreshold = 0.90
  pornThreshold = 0.81
  sexyThreshold = 0.85
  overallThreshold = 0.90
```
