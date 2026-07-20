package avito

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const apiBaseURL = "https://api.avito.ru"

type Client struct {
	clientID     string
	clientSecret string
	userID       int64
	httpClient   *http.Client

	mu             sync.Mutex
	accessToken    string
	tokenExpiresAt time.Time
}

func NewClient(clientID, clientSecret string, userID int64) *Client {
	return &Client{
		clientID:     strings.TrimSpace(clientID),
		clientSecret: strings.TrimSpace(clientSecret),
		userID:       userID,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.clientID != "" && c.clientSecret != "" && c.userID > 0
}

func (c *Client) UserID() int64 {
	if c == nil {
		return 0
	}
	return c.userID
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func (c *Client) ensureAccessToken(ctx context.Context) (string, error) {
	if !c.Enabled() {
		return "", fmt.Errorf("avito client is not configured")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.accessToken != "" && time.Now().Before(c.tokenExpiresAt.Add(-2*time.Minute)) {
		return c.accessToken, nil
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiBaseURL+"/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("avito token error: status=%d body=%s", res.StatusCode, truncate(string(body), 300))
	}

	var payload tokenResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}
	if strings.TrimSpace(payload.AccessToken) == "" {
		return "", fmt.Errorf("avito token response has empty access_token")
	}

	expiresIn := payload.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 86400
	}
	c.accessToken = payload.AccessToken
	c.tokenExpiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)
	return c.accessToken, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, query url.Values, requestBody any, out any) error {
	token, err := c.ensureAccessToken(ctx)
	if err != nil {
		return err
	}

	endpoint := apiBaseURL + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	var bodyReader io.Reader
	if requestBody != nil {
		raw, marshalErr := json.Marshal(requestBody)
		if marshalErr != nil {
			return marshalErr
		}
		bodyReader = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if requestBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusUnauthorized {
		c.mu.Lock()
		c.accessToken = ""
		c.tokenExpiresAt = time.Time{}
		c.mu.Unlock()
		return fmt.Errorf("avito unauthorized: %s", truncate(string(body), 300))
	}
	if res.StatusCode >= 300 {
		return fmt.Errorf("avito api error: status=%d body=%s", res.StatusCode, truncate(string(body), 400))
	}
	if out == nil || len(body) == 0 {
		return nil
	}
	return json.Unmarshal(body, out)
}

type ChatListResponse struct {
	Chats []Chat `json:"chats"`
	Meta  struct {
		HasMore bool `json:"has_more"`
	} `json:"meta"`
}

type Chat struct {
	ID          string       `json:"id"`
	Created     int64        `json:"created"`
	Updated     int64        `json:"updated"`
	Users       []ChatUser   `json:"users"`
	Context     ChatContext  `json:"context"`
	LastMessage *ChatMessage `json:"last_message"`
}

type ChatUser struct {
	ID                int64              `json:"id"`
	Name              string             `json:"name"`
	PublicUserProfile *PublicUserProfile `json:"public_user_profile"`
}

type PublicUserProfile struct {
	UserID int64  `json:"user_id"`
	URL    string `json:"url"`
	Avatar *struct {
		Default string            `json:"default"`
		Images  map[string]string `json:"images"`
	} `json:"avatar"`
}

type ChatContext struct {
	Type  string `json:"type"`
	Value struct {
		ID    int64  `json:"id"`
		Title string `json:"title"`
		URL   string `json:"url"`
	} `json:"value"`
}

type ChatMessage struct {
	ID        string `json:"id"`
	AuthorID  int64  `json:"author_id"`
	Created   int64  `json:"created"`
	Type      string `json:"type"`
	Direction string `json:"direction"`
	IsRead    bool   `json:"isRead"`
	Content   struct {
		Text  string `json:"text"`
		Image *struct {
			Sizes map[string]string `json:"sizes"`
		} `json:"image"`
	} `json:"content"`
}

type UploadFile struct {
	Filename    string
	ContentType string
	Data        []byte
}

type MessagesResponse struct {
	Messages []ChatMessage `json:"messages"`
	Meta     struct {
		HasMore bool `json:"has_more"`
	} `json:"meta"`
}

func (c *Client) GetChats(ctx context.Context, limit, offset int) (ChatListResponse, error) {
	if limit <= 0 {
		limit = 50
	}
	query := url.Values{}
	query.Set("limit", strconv.Itoa(limit))
	query.Set("offset", strconv.Itoa(offset))

	var out ChatListResponse
	path := fmt.Sprintf("/messenger/v2/accounts/%d/chats", c.userID)
	err := c.doJSON(ctx, http.MethodGet, path, query, nil, &out)
	return out, err
}

func (c *Client) GetChat(ctx context.Context, chatID string) (Chat, error) {
	var out Chat
	path := fmt.Sprintf("/messenger/v2/accounts/%d/chats/%s", c.userID, url.PathEscape(chatID))
	err := c.doJSON(ctx, http.MethodGet, path, nil, nil, &out)
	return out, err
}

func (c *Client) GetMessages(ctx context.Context, chatID string, limit, offset int) (MessagesResponse, error) {
	if limit <= 0 {
		limit = 50
	}
	query := url.Values{}
	query.Set("limit", strconv.Itoa(limit))
	query.Set("offset", strconv.Itoa(offset))

	var out MessagesResponse
	path := fmt.Sprintf("/messenger/v3/accounts/%d/chats/%s/messages/", c.userID, url.PathEscape(chatID))
	err := c.doJSON(ctx, http.MethodGet, path, query, nil, &out)
	return out, err
}

func (c *Client) SendTextMessage(ctx context.Context, chatID, text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return fmt.Errorf("message text is required")
	}
	body := map[string]any{
		"message": map[string]string{"text": text},
		"type":    "text",
	}
	path := fmt.Sprintf("/messenger/v1/accounts/%d/chats/%s/messages", c.userID, url.PathEscape(chatID))
	return c.doJSON(ctx, http.MethodPost, path, nil, body, nil)
}

func (c *Client) UploadImage(ctx context.Context, file UploadFile) (imageID string, sizes map[string]string, err error) {
	if !c.Enabled() {
		return "", nil, fmt.Errorf("avito client is not configured")
	}
	if len(file.Data) == 0 {
		return "", nil, fmt.Errorf("image file is empty")
	}
	if len(file.Data) > 24<<20 {
		return "", nil, fmt.Errorf("image is too large (max 24 MB)")
	}

	token, err := c.ensureAccessToken(ctx)
	if err != nil {
		return "", nil, err
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	filename := strings.TrimSpace(file.Filename)
	if filename == "" {
		filename = "image.jpg"
	}
	part, createErr := writer.CreateFormFile("uploadfile[]", filepath.Base(filename))
	if createErr != nil {
		return "", nil, createErr
	}
	if _, writeErr := part.Write(file.Data); writeErr != nil {
		return "", nil, writeErr
	}
	if closeErr := writer.Close(); closeErr != nil {
		return "", nil, closeErr
	}

	path := fmt.Sprintf("/messenger/v1/accounts/%d/uploadImages", c.userID)
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, apiBaseURL+path, &body)
	if reqErr != nil {
		return "", nil, reqErr
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, doErr := c.httpClient.Do(req)
	if doErr != nil {
		return "", nil, doErr
	}
	defer res.Body.Close()

	raw, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		return "", nil, readErr
	}
	if res.StatusCode == http.StatusUnauthorized {
		c.mu.Lock()
		c.accessToken = ""
		c.tokenExpiresAt = time.Time{}
		c.mu.Unlock()
		return "", nil, fmt.Errorf("avito unauthorized: %s", truncate(string(raw), 300))
	}
	if res.StatusCode >= 300 {
		return "", nil, fmt.Errorf("avito upload error: status=%d body=%s", res.StatusCode, truncate(string(raw), 400))
	}

	var payload map[string]map[string]string
	if unmarshalErr := json.Unmarshal(raw, &payload); unmarshalErr != nil {
		return "", nil, fmt.Errorf("invalid upload response: %w", unmarshalErr)
	}
	for id, imageSizes := range payload {
		if strings.TrimSpace(id) == "" {
			continue
		}
		return id, imageSizes, nil
	}
	return "", nil, fmt.Errorf("avito upload response has no image id")
}

func (c *Client) SendImageMessage(ctx context.Context, chatID, imageID string) error {
	imageID = strings.TrimSpace(imageID)
	if imageID == "" {
		return fmt.Errorf("image_id is required")
	}
	body := map[string]any{"image_id": imageID}
	path := fmt.Sprintf("/messenger/v1/accounts/%d/chats/%s/messages/image", c.userID, url.PathEscape(chatID))
	return c.doJSON(ctx, http.MethodPost, path, nil, body, nil)
}

func MessageText(msg ChatMessage) string {
	text := strings.TrimSpace(msg.Content.Text)
	if text != "" {
		return text
	}
	if msg.Content.Image != nil && len(msg.Content.Image.Sizes) > 0 {
		if url := PreferredImageURL(msg.Content.Image.Sizes); url != "" {
			return url
		}
	}
	if strings.TrimSpace(msg.Type) != "" && msg.Type != "text" {
		return "[" + msg.Type + "]"
	}
	return ""
}

func PreferredImageURL(sizes map[string]string) string {
	preferredKeys := []string{"1280x960", "640x480", "320x240", "140x105"}
	for _, key := range preferredKeys {
		if value := strings.TrimSpace(sizes[key]); value != "" {
			return value
		}
	}
	best := ""
	for _, value := range sizes {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if best == "" || len(value) > len(best) {
			best = value
		}
	}
	return best
}

func (c *Client) SubscribeWebhook(ctx context.Context, webhookURL string) error {
	body := map[string]string{"url": strings.TrimSpace(webhookURL)}
	return c.doJSON(ctx, http.MethodPost, "/messenger/v3/webhook", nil, body, nil)
}

func (c *Client) PeerFromChat(chat Chat) (peerID int64, nickname, avatarURL string) {
	for _, user := range chat.Users {
		if user.ID == c.userID {
			continue
		}
		nickname = strings.TrimSpace(user.Name)
		peerID = user.ID
		if user.PublicUserProfile != nil && user.PublicUserProfile.Avatar != nil {
			avatarURL = strings.TrimSpace(user.PublicUserProfile.Avatar.Default)
			if avatarURL == "" && user.PublicUserProfile.Avatar.Images != nil {
				avatarURL = strings.TrimSpace(user.PublicUserProfile.Avatar.Images["128x128"])
			}
		}
		return peerID, nickname, avatarURL
	}
	return 0, "", ""
}

func truncate(value string, max int) string {
	if max <= 0 || len(value) <= max {
		return value
	}
	return value[:max] + "…"
}
