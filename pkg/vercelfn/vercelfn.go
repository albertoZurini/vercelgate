package vercelfn

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/khanakia/vercelgate/pkg/accountstore"
	"github.com/khanakia/vercelgate/pkg/logger"
	"github.com/khanakia/vercelgate/pkg/vercelapi"
	"github.com/khanakia/vercelgate/pkg/vercelutil"
)

func SyncAuthJson() error {
	logger.Verbose("SyncAuthJson()")

	authJsonFile, err := vercelutil.AuthJsonFile()
	if err != nil {
		return err
	}

	logger.Debug("reading auth file: %s", authJsonFile)

	rawBytes, err := os.ReadFile(authJsonFile)
	if err != nil {
		return err
	}

	var authConfig vercelutil.AuthConfig
	if err := json.Unmarshal(rawBytes, &authConfig); err != nil {
		return err
	}

	logger.Debug("fetching user from Vercel API")
	user, err := vercelapi.GetUser(authConfig.Token)
	if err != nil {
		return err
	}

	displayName := user.Username
	if len(user.Name) > 0 {
		displayName = user.Name
	}
	name := fmt.Sprintf("%s (%s)", displayName, user.Email)

	logger.Debug("storing account: %s", name)
	return accountstore.AddOrUpdate(name, rawBytes)
}
