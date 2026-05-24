package vercelapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/khanakia/vercelgate/pkg/logger"
)

// baseURL is the Vercel API root; overridden in tests.
var baseURL = "https://api.vercel.com"

func GetUser(token string) (*User, error) {
	logger.Verbose("GetUser(token=%s)", logger.MaskToken(token))

	if len(token) == 0 {
		return nil, errors.New("token is empty")
	}

	url := baseURL + "/v2/user"
	logger.Debug("GET %s", url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
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

func GetTeams(token string) ([]Team, error) {
	logger.Verbose("GetTeams(token=%s)", logger.MaskToken(token))

	if len(token) == 0 {
		return nil, errors.New("token is empty")
	}

	url := baseURL + "/v2/teams"
	logger.Debug("GET %s", url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	logger.Debug("GET %s -> %d", url, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.New("unable to fetch package info")
	}

	var record *GetTeamsResponse
	err = json.Unmarshal(body, &record)
	if err != nil {
		return nil, errors.New("unable to parse json")
	}

	if len(record.Error.Code) > 0 {
		logger.Debug("API error: code=%s message=%s", record.Error.Code, record.Error.Message)
		return nil, errors.New(record.Error.Message)
	}

	logger.Debug("got %d team(s)", len(record.Teams))
	return record.Teams, nil
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

type Team struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	Avatar    any       `json:"avatar"`
	CreatedAt int64     `json:"createdAt"`
	Created   time.Time `json:"created"`
}

type Pagination struct {
	Count int   `json:"count"`
	Next  int64 `json:"next"`
	Prev  int64 `json:"prev"`
}

type GetTeamsResponse struct {
	Teams      []Team     `json:"teams"`
	Pagination Pagination `json:"pagination"`
	Error      Error      `json:"error,omitempty"`
}
