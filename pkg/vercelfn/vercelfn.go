package vercelfn

import (
	"context"

	"github.com/khanakia/vercelgate/pkg/entdb"
	"github.com/khanakia/vercelgate/pkg/logger"
	"github.com/khanakia/vercelgate/pkg/vercelapi"
	"github.com/khanakia/vercelgate/pkg/vercelutil"
)

// Sycn user and team to database from auth.json file
func SyncAuthJson() error {
	logger.Verbose("SyncAuthJson()")

	authJsonFile, err := vercelutil.AuthJsonFile()
	if err != nil {
		return err
	}

	logger.Debug("reading auth file: %s", authJsonFile)

	authConfig, err := vercelutil.ParseAuthFile(authJsonFile)
	if err != nil {
		return err
	}

	logger.Debug("fetching user from Vercel API")
	user, err := vercelapi.GetUser(authConfig.Token)
	if err != nil {
		return err
	}

	ctx := context.Background()

	logger.Debug("upserting user id=%s email=%s", user.ID, user.Email)
	userID, err := entdb.Client().User.Create().
		SetID(user.ID).
		SetEmail(user.Email).
		SetName(user.Name).
		SetUsername(user.Username).
		SetToken(authConfig.Token).
		OnConflict().
		UpdateNewValues().
		ID(ctx)
	if err != nil {
		return err
	}

	logger.Debug("fetching teams from Vercel API")
	teams, err := vercelapi.GetTeams(authConfig.Token)
	if err != nil {
		return err
	}

	for _, team := range teams {
		logger.Debug("upserting team id=%s name=%s slug=%s", team.ID, team.Name, team.Slug)
		_, err := entdb.Client().Team.Create().
			SetID(team.ID).
			SetUserID(userID).
			SetName(team.Name).
			SetSlug(team.Slug).
			OnConflict().
			UpdateNewValues().
			ID(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
