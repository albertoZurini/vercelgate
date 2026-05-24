package accountstore

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/khanakia/vercelgate/pkg/vercelutil"
)

const fileName = "vercelgate_accounts.json"

type Account struct {
	Name string          `json:"name"`
	Data json.RawMessage `json:"data"`
}

func FilePath() (string, error) {
	globalPath, err := vercelutil.GetGlobalPathConfig()
	if err != nil {
		return "", err
	}
	return filepath.Join(globalPath, fileName), nil
}

func Load() ([]Account, error) {
	path, err := FilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Account{}, nil
		}
		return nil, err
	}

	var accounts []Account
	if err := json.Unmarshal(data, &accounts); err != nil {
		return nil, err
	}
	return accounts, nil
}

func Save(accounts []Account) error {
	path, err := FilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(accounts, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func AddOrUpdate(name string, data json.RawMessage) error {
	accounts, err := Load()
	if err != nil {
		return err
	}

	for i, a := range accounts {
		if a.Name == name {
			accounts[i].Data = data
			return Save(accounts)
		}
	}

	accounts = append(accounts, Account{Name: name, Data: data})
	return Save(accounts)
}

func All() ([]Account, error) {
	return Load()
}

func Clear() error {
	return Save([]Account{})
}
