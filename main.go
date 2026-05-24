package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/khanakia/vercelgate/gen/ent"
	"github.com/khanakia/vercelgate/gen/ent/team"
	"github.com/khanakia/vercelgate/pkg/constants"
	"github.com/khanakia/vercelgate/pkg/entcfn"
	"github.com/khanakia/vercelgate/pkg/entdb"
	"github.com/khanakia/vercelgate/pkg/logger"
	"github.com/khanakia/vercelgate/pkg/utils"
	"github.com/khanakia/vercelgate/pkg/vercelfn"
	"github.com/khanakia/vercelgate/pkg/vercelutil"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	_ "github.com/mattn/go-sqlite3"
)

var version = "dev"

var (
	debugFlag   bool
	verboseFlag bool
)

func main() {
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "enable debug logging")
	rootCmd.PersistentFlags().BoolVar(&verboseFlag, "verbose", false, "log function arguments (requires --debug)")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(resetCmd)
	rootCmd.AddCommand(switchCmd)
	rootCmd.AddCommand(switchTeamCmd)
	rootCmd.AddCommand(pathCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:     "vercelgate",
	Version: version,
	Short:   "Make vercel cli more powerful by adding the ability to switch between multiple accounts.",
	Long:    `You can swithc between multiple accounts without having relogin and logout.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.DebugEnabled = debugFlag
		// --verbose implies --debug
		if verboseFlag {
			logger.DebugEnabled = true
			logger.VerboseEnabled = true
		}
		logger.Debug("debug logging enabled")
		logger.Verbose("verbose logging enabled (function arguments will be logged)")
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Run command `vercelgate --help` for more information`")
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Run this command very first time",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("running init command")
		globalPath, err := vercelutil.GetGlobalPathConfig()
		if err != nil {
			log.Fatal(err)

			return
		}

		dbPath := filepath.Join(globalPath, constants.DB_FILE_NAME)
		logger.Debug("checking for existing DB at %s", dbPath)
		err = utils.IsFileExists(dbPath)
		if err == nil {
			fmt.Println("was initialized already")
			return
		}

		logger.Debug("running schema migration")
		err = entcfn.Migrate()
		if err != nil {
			log.Fatal(err)

			return
		}
		fmt.Println("vercelgate initialized successfully")
	},
}

var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switch between account",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("running switch command")
		err := SwitchCmd(false)

		if err != nil {
			log.Fatal(err)

			return
		}
	},
}

var switchTeamCmd = &cobra.Command{
	Use:   "switchteam",
	Short: "Switch between account and teams",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("running switchteam command")
		err := SwitchCmd(true)

		if err != nil {
			log.Fatal(err)

			return
		}
	},
}

func SwitchCmd(switchTeam bool) error {
	logger.Verbose("SwitchCmd(switchTeam=%v)", switchTeam)

	user, err := promptGetUser()
	if err != nil {
		return err
	}

	logger.Debug("selected user: %s (%s)", user.Name, user.Email)

	err = vercelutil.SetAuthToken(user.Token)
	if err != nil {
		return err
	}

	fmt.Printf("Switched to user %s\n", user.Name)

	if switchTeam {
		team, err := promptGetTeam(user.ID)
		if err != nil {
			return err
		}

		logger.Debug("selected team: %s (%s)", team.Name, team.ID)

		err = vercelutil.SetCurrentTeam(team.ID)
		if err != nil {
			return err
		}

		fmt.Printf("Switched to team %s\n", team.Name)
	} else {
		logger.Debug("clearing currentTeam (no team switch)")
		vercelutil.DeleteCurrentTeam()
	}

	return nil
}

func promptGetTeam(userID string) (*ent.Team, error) {
	ctx := context.Background()

	items, err := entdb.Client().Team.Query().Where(team.UserID(userID)).All(ctx)

	if err != nil {
		return nil, err
	}

	itemsList := []string{}

	for _, user := range items {
		itemsList = append(itemsList, user.Name)
	}

	prompt := promptui.Select{
		Label: "Select Team",
		Items: itemsList,
	}

	index, _, err := prompt.Run()

	if err != nil {
		return nil, fmt.Errorf("Prompt failed %v\n", err)
	}

	return items[index], nil
}

func promptGetUser() (*ent.User, error) {
	ctx := context.Background()

	users, err := entdb.Client().User.Query().All(ctx)

	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("no accounts synced yet. please sync first using `vercelgate sync`")
	}

	usersList := []string{}

	for _, user := range users {
		name := user.Name
		if len(name) == 0 {
			name = user.Username
		}
		usersList = append(usersList, fmt.Sprintf("%s (%s)", name, user.Email))
	}

	prompt := promptui.Select{
		Label: "Select Account",
		Items: usersList,
	}

	index, _, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return nil, err
	}

	return users[index], nil
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync current logged in account",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("running sync command")
		err := vercelfn.SyncAuthJson()
		if err != nil {
			log.Fatal(err)
			return
		}

		fmt.Println("synced successfully")
	},
}

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Run this to add new vercel client account",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("running new command")
		err := NewAccountCmd()
		if err != nil {
			log.Fatal(err)
			return
		}

		fmt.Println("you can now add new account using `vercel login` and then run `vercelgate sync` again")
	},
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset vercelgate state and will delete all accounts",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("running reset command")
		err := ResetCmd()
		if err != nil {
			log.Fatal(err)
			return
		}

		fmt.Println("state reset was successful")
	},
}

var pathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show Vercel global configuration path",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("running path command")
		globalPath, err := vercelutil.GetGlobalPathConfig()
		if err != nil {
			log.Fatal(err)
			return
		}
		fmt.Printf("Vercel global configuration path: %s\n", globalPath)
	},
}

func NewAccountCmd() error {
	filePath, err := vercelutil.AuthJsonFile()
	if err != nil {
		return err
	}

	logger.Debug("removing auth.json at %s", filePath)
	err = os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("failed to remove auth.json file: %w", err)
	}
	return nil
}

func ResetCmd() error {
	logger.Debug("deleting all users and teams from DB")
	ctx := context.Background()
	_, err := entdb.Client().User.Delete().Exec(ctx)
	if err != nil {
		return err
	}
	_, err = entdb.Client().Team.Delete().Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}
