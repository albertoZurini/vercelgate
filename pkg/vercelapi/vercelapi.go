package vercelapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/albertoZurini/vercelgate/pkg/logger"
)

// baseURL is the Vercel API root; overridden in tests.
var baseURL = "https://api.vercel.com"

var httpClient = &http.Client{Timeout: 15 * time.Second}

func GetUser(token string) (*User, error) {
	logger.Verbose("GetUser(token=%s)", logger.MaskToken(token))

	if len(token) == 0 {
		return nil, errors.New("token is empty")
	}

	url := baseURL + "/v2/user"
	logger.Debug("GET %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	logger.Debug("GET %s -> %d", url, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.New("unable to fetch package info")
	}

	var record *GetUserResponse
	err = json.Unmarshal(body, &record)
	if err != nil {
		return nil, errors.New("unable to parse json")
	}

	if len(record.Error.Code) > 0 {
		logger.Debug("API error: code=%s message=%s", record.Error.Code, record.Error.Message)
		return nil, errors.New(record.Error.Message)
	}

	logger.Debug("got user: id=%s email=%s username=%s", record.User.ID, record.User.Email, record.User.Username)
	return &record.User, nil
}

type GetUserResponse struct {
	User  User  `json:"user"`
	Error Error `json:"error,omitempty"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type User struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Username      string `json:"username"`
	Avatar        any    `json:"avatar"`
	DefaultTeamID string `json:"defaultTeamId"`
	Version       string `json:"version"`
	CreatedAt     int64  `json:"createdAt"`
}

