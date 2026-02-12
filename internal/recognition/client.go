package recognition

import (
	"fmt"

	"github.com/mystaline/myrics-overlay/pkg/models"
)

// Client handles music recognition via ACRCloud API
type Client struct {
	accessKey string
	secretKey string
	host      string
}

// NewClient creates a new ACRCloud client
func NewClient(accessKey, secretKey, host string) *Client {
	return &Client{
		accessKey: accessKey,
		secretKey: secretKey,
		host:      host,
	}
}

// Recognize identifies a song from audio data
func (c *Client) Recognize(audioData []byte) (*models.SongInfo, error) {
	// TODO: Step 4 - Implement ACRCloud API integration

	// ACRCloud API documentation:
	// https://docs.acrcloud.com/reference/identification-api

	return nil, fmt.Errorf("not implemented")
}
