//go:build windows

package recognition

import (
	"fmt"

	"github.com/mystaline/myrics-overlay/pkg/models"
)

// Client is a stub on Windows.
// The ACRCloud SDK requires a Linux shared library (.so) and is not available on Windows.
// Song detection on Windows is handled by SMTC instead.
type Client struct{}

func NewClient(_, _, _ string) *Client { return &Client{} }

func (c *Client) Recognize(_ []byte) (*models.SongInfo, error) {
	return nil, fmt.Errorf("ACRCloud not available on Windows; use SMTC for song detection")
}
