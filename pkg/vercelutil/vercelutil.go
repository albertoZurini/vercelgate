package vercelutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/khanakia/vercelgate/pkg/jsonupdate"
	"github.com/khanakia/vercelgate/pkg/logger"
	"github.com/khanakia/vercelgate/pkg/utils"

	"github.com/adrg/xdg"
)

var (
	authJsonFileName = "auth.json"
)

func SetAuthToken(token string) error {
	logger.Verbose("SetAuthToken(token=%s)", logger.MaskToken(token))

	filePath, err := AuthJsonFile()
	if err != nil {
		return err
	}

	logger.Debug("writing token to %s", filePath)

	fileBytes, err := utils.OpenFile(filePath)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	jsonupd := jsonupdate.NewJsonUpdate(string(fileBytes))

	jsonupd.Set("token", token)

	err = os.WriteFile(filePath, []byte(jsonupd.String()), 0644)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	return nil
}

func SetCurrentTeam(teamID string) error {
	logger.Verbose("SetCurrentTeam(teamID=%s)", teamID)

	filePath, err := ConfigJsonFile()
	if err != nil {
		return err
	}

	logger.Debug("writing currentTeam to %s", filePath)

	fileBytes, err := utils.OpenFile(filePath)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	jsonupd := jsonupdate.NewJsonUpdate(string(fileBytes))

	jsonupd.Set("currentTeam", teamID)

	err = os.WriteFile(filePath, []byte(jsonupd.Pretty()), 0644)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	return nil
}

func DeleteCurrentTeam() error {
	logger.Verbose("DeleteCurrentTeam()")

	filePath, err := ConfigJsonFile()
	if err != nil {
		return err
	}

	logger.Debug("removing currentTeam from %s", filePath)

	fileBytes, err := utils.OpenFile(filePath)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	jsonupd := jsonupdate.NewJsonUpdate(string(fileBytes))

	jsonupd.Deleete("currentTeam")

	err = os.WriteFile(filePath, []byte(jsonupd.String()), 0644)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	return nil
}

func ParseAuthFile(path string) (*AuthConfig, error) {
	logger.Verbose("ParseAuthFile(path=%s)", path)

	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var authConfig AuthConfig
	err = json.Unmarshal(fileBytes, &authConfig)
	if err != nil {
		return nil, err
	}

	logger.Debug("parsed auth file, token present: %v", len(authConfig.Token) > 0)
	return &authConfig, nil
}

type AuthConfig struct {
	Token string `json:"token"`
}

func AuthJsonFile() (string, error) {
	globalPath, err := GetGlobalPathConfig()
	if err != nil {
		return "", err
	}
	return filepath.Join(globalPath, authJsonFileName), nil
}

func ConfigJsonFile() (string, error) {
	globalPath, err := GetGlobalPathConfig()
	if err != nil {
		return "", err
	}
	return filepath.Join(globalPath, "config.json"), nil
}

func GetGlobalPathConfig() (string, error) {
	dirname := "com.vercel.cli"

	dirs := append(xdg.ConfigDirs, xdg.ConfigHome)

	logger.Debug("searching for vercel config dir in %d candidate path(s)", len(dirs))

	for _, datadir := range dirs {
		dirPath := filepath.Join(datadir, dirname)
		logger.Verbose("GetGlobalPathConfig: checking %s", dirPath)

		info, err := os.Stat(dirPath)
		if err != nil {
			continue
		}

		if info.IsDir() {
			logger.Debug("found vercel config dir: %s", dirPath)
			return dirPath, nil
		}
	}

	return "", errors.New("global path config not found")
}
