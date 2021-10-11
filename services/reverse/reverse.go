package reverse

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gitlab.snapp.ir/Map/sdk/smapp-sdk-go/config"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Interface consists of functions of different functionalities of a reverse geocode service. there are two implementation of this service.
// one for mocking and one for production usage.
type Interface interface {
	// GetComponents receives `lat`,`lon` as a location and CallOptions and returns Component s of address of location given.
	GetComponents(lat, lon float64, options CallOptions) ([]Component, error)
	// GetDisplayName receives `lat`,`lon`` as a location and CallOptions and returns a string as address of given location.
	GetDisplayName(lat, lon float64, options CallOptions) (string, error)
	// GetComponentsWithContext is like GetComponents, but with context.Context support.
	GetComponentsWithContext(ctx context.Context, lat, lon float64, options CallOptions) ([]Component, error)
	// GetDisplayNameWithContext is like GetDisplayName, but with context.Context support.
	GetDisplayNameWithContext(ctx context.Context, lat, lon float64, options CallOptions) (string, error)
}

type Version string

const (
	V1 Version = "v1"

	Lat       = "lat"
	Lon       = "lon"
	Lang      = "language"
	ZoomLevel = "zoom"
	Type      = "type"
	Display   = "display"

	OKStatus    = "OK"
	ErrorStatus = "ERROR"
)

// Client is the main implementation of Interface for search service
type Client struct {
	cfg        *config.Config
	url        string
	httpClient http.Client
}

// Force Client to implement Interface at compile time
var _ Interface = (*Client)(nil)

// GetComponents receives `lat`,`lon` as a location and CallOptions and returns Component s of address of location given.
func (c *Client) GetComponents(lat, lon float64, options CallOptions) ([]Component, error) {
	return c.GetComponentsWithContext(context.Background(), lat, lon, options)
}

// GetDisplayName receives `lat`,`lon`` as a location and CallOptions and returns a string as address of given location.
func (c *Client) GetDisplayName(lat, lon float64, options CallOptions) (string, error) {
	return c.GetDisplayNameWithContext(context.Background(), lat, lon, options)
}

// GetComponentsWithContext is like GetComponents, but with context.Context support.
func (c *Client) GetComponentsWithContext(ctx context.Context, lat, lon float64, options CallOptions) ([]Component, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return nil, errors.New("smapp reverse geo-code: could not create request. err: " + err.Error())
	}

	params := url.Values{}

	params.Set(Lat, fmt.Sprintf("%f", lat))
	params.Set(Lon, fmt.Sprintf("%f", lon))
	if options.UseLanguage {
		params.Set(Lang, string(options.Language))
	}

	if options.UseZoomLevel {
		params.Set(ZoomLevel, strconv.Itoa(options.ZoomLevel))
	}

	if options.UseResponseType {
		params.Set(Type, string(options.ResponseType))
	}
	params.Set(Display, "false")

	if c.cfg.APIKeySource == config.HeaderSource {
		req.Header.Set(c.cfg.APIKeyName, c.cfg.APIKey)
	} else if c.cfg.APIKeySource == config.QueryParamSource {
		params.Set(c.cfg.APIKeyName, c.cfg.APIKey)
	} else {
		return nil, fmt.Errorf("smapp reverse geo-code: invalid api key source: %s", string(c.cfg.APIKeySource))
	}

	for key, val := range options.Headers {
		req.Header.Set(key, val)
	}

	req.URL.RawQuery = params.Encode()

	response, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("smapp reverse geo-code: could not make a request due to this error: %s", err.Error())
	}

	defer func() {
		_, _ = io.Copy(ioutil.Discard, response.Body)
		_ = response.Body.Close()
	}()

	if response.StatusCode == http.StatusOK {
		resp := struct {
			Status string `json:"status"`
			Result struct {
				Components []Component `json:"components"`
			} `json:"result"`
		}{}

		err := json.NewDecoder(response.Body).Decode(&resp)
		if err != nil {
			return nil, fmt.Errorf("smapp reverse geo-code: could not serialize response due to: %s", err.Error())
		}

		if resp.Status != OKStatus {
			return nil, errors.New("smapp reverse geo-code: status of request is not OK")
		}

		return resp.Result.Components, nil
	}

	return nil, fmt.Errorf("smapp reverse geo-code: non 200 status: %d", response.StatusCode)
}

// GetDisplayNameWithContext is like GetDisplayName, but with context.Context support.
func (c *Client) GetDisplayNameWithContext(ctx context.Context, lat, lon float64, options CallOptions) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return "", errors.New("smapp reverse geo-code: could not create request. err: " + err.Error())
	}

	params := url.Values{}

	params.Set(Lat, fmt.Sprintf("%f", lat))
	params.Set(Lon, fmt.Sprintf("%f", lon))

	if options.UseLanguage {
		params.Set(Lang, string(options.Language))
	}

	if options.UseZoomLevel {
		params.Set(ZoomLevel, strconv.Itoa(options.ZoomLevel))
	}

	if options.UseResponseType {
		params.Set(Type, string(options.ResponseType))
	}

	params.Set(Display, "true")

	if c.cfg.APIKeySource == config.HeaderSource {
		req.Header.Set(c.cfg.APIKeyName, c.cfg.APIKey)
	} else if c.cfg.APIKeySource == config.QueryParamSource {
		params.Set(c.cfg.APIKeyName, c.cfg.APIKey)
	} else {
		return "", fmt.Errorf("smapp reverse geo-code: invalid api key source: %s", string(c.cfg.APIKeySource))
	}

	for key, val := range options.Headers {
		req.Header.Set(key, val)
	}

	req.URL.RawQuery = params.Encode()

	response, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("smapp reverse geo-code: could not make a request due to this error: %s", err.Error())
	}

	defer func() {
		_, _ = io.Copy(ioutil.Discard, response.Body)
		_ = response.Body.Close()
	}()

	if response.StatusCode == http.StatusOK {
		resp := struct {
			Status string `json:"status"`
			Result struct {
				DisplayName string `json:"displayName"`
			} `json:"result"`
		}{}

		err := json.NewDecoder(response.Body).Decode(&resp)
		if err != nil {
			return "", fmt.Errorf("smapp reverse geo-code: could not serialize response due to: %s", err.Error())
		}

		if resp.Status != OKStatus {
			return "", errors.New("smapp reverse geo-code: status of request is not OK")
		}

		return resp.Result.DisplayName, nil
	}

	return "", fmt.Errorf("smapp reverse geo-code: non 200 status: %d", response.StatusCode)
}

// NewReverseClient is the constructor of reverse geocode client.
func NewReverseClient(cfg *config.Config, version Version, timeout time.Duration, opts ...ConstructorOption) (*Client, error) {
	client := &Client{
		cfg: cfg,
		url: getReverseDefaultURL(cfg, version),
		httpClient: http.Client{
			Timeout: timeout,
		},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

func getReverseDefaultURL(cfg *config.Config, version Version) string {
	baseURL := strings.TrimRight(cfg.APIBaseURL, "/")
	return fmt.Sprintf("%s/reverse/%s", baseURL, version)
}