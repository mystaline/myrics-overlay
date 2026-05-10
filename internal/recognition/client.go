//go:build !windows

package recognition

import (
	"github.com/acrcloud/acrcloud_sdk_golang/acrcloud"
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
// Recognize identifies a song from audio data
func (c *Client) Recognize(audioData []byte) (*models.SongInfo, error) {
	// Initialize recognizer with options
	// Note: We use map config as per SDK examples
	config := map[string]string{
		"access_key":     c.accessKey,
		"secret_key":     c.secretKey,
		"host":           c.host,
		"recognize_type": acrcloud.ACR_OPT_REC_AUDIO, // Recognize by audio
		"debug":          "false",
		"timeout":        "10", // Seconds
	}

	recognizer := acrcloud.NewRecognizer(config)

	// Recognize by audio buffer
	// We don't need 'start_seconds_begin' as we captured a snippet specifically for recognition
	// user_params can be nil (map[string]string)
	result := recognizer.Recognize(audioData, nil)

	// The SDK usage seems to return just the result string, errors are embedded in the JSON?
	// Or maybe it does return error but my LSP was wrong?
	// Checking source code of SDK: func (r *Recognizer) Recognize(pcm []byte) string
	// It only returns string.

	// Parse the JSON result string into our model
	return models.ParseACRCloudResponse(result)
}
