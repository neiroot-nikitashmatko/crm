package max

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const apiBaseURL = "https://platform-api2.max.ru"

type Client struct {
	token      string
	chatID     int64
	httpClient *http.Client
}

func NewClient(token string, chatID int64, tlsInsecure bool) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if tlsInsecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return &Client{
		token:  strings.TrimSpace(token),
		chatID: chatID,
		httpClient: &http.Client{
			Timeout:   20 * time.Second,
			Transport: transport,
		},
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.token != "" && c.chatID != 0
}

func (c *Client) SendText(ctx context.Context, text string) error {
	return c.SendFormattedText(ctx, text, "")
}

func (c *Client) SendFormattedText(ctx context.Context, text string, format string) error {
	if !c.Enabled() {
		return fmt.Errorf("max client is not configured")
	}

	payload := map[string]string{"text": text}
	if format = strings.TrimSpace(format); format != "" {
		payload["format"] = format
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	endpoint := apiBaseURL + "/messages?" + url.Values{
		"chat_id": {strconv.FormatInt(c.chatID, 10)},
	}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "certificate") {
			return fmt.Errorf("%w (локально на macOS: MAX_TLS_INSECURE=true)", err)
		}
		return err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("max send failed: status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	return nil
}
