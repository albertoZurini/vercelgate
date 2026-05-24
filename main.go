package main

import (
	"fmt"
	"log"
	"os"

	"github.com/albertoZurini/vercelgate/pkg/accountstore"
	"github.com/albertoZurini/vercelgate/pkg/logger"
	"github.com/albertoZurini/vercelgate/pkg/vercelfn"
	"github.com/albertoZurini/vercelgate/pkg/vercelutil"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
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
	Long:    `You can switch between multiple accounts without having to relogin and logout.`,
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
		path, err := accountstore.FilePath()
		if err != nil {
			log.Fatal(err)
			return
		}

		logger.Debug("checking for existing accounts file at %s", path)
		if _, err := os.Stat(path); err == nil {
			fmt.Println("was initialized already")
			return
		}

		logger.Debug("creating empty accounts file")
		if err := accountstore.Clear(); err != nil {
			log.Fatal(err)
			return
		}
		fmt.Println("vercelgate initialized successfully")
	},
}

var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switch between accounts",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("running switch command")
		err := SwitchCmd()

		if err != nil {
			log.Fatal(err)

			return
		}
	},
}

func SwitchCmd() error {
	logger.Verbose("SwitchCmd()")

	account, err := promptGetUser()
	if err != nil {
		return err
	}

	logger.Debug("selected account: %s", account.Name)

	if err := vercelutil.RestoreAuthJson(account.Data); err != nil {
		return err
	}

	fmt.Printf("Switched to %s\n", account.Name)

	logger.Debug("clearing currentTeam")
	if err := vercelutil.DeleteCurrentTeam(); err != nil {
		return err
	}

	return nil
}

func promptGetUser() (accountstore.Account, error) {
	accounts, err := accountstore.All()
	if err != nil {
		return accountstore.Account{}, err
	}

	if len(accounts) == 0 {
		return accountstore.Account{}, fmt.Errorf("no accounts synced yet. please sync first using `vercelgate sync`")
	}

	items := make([]string, len(accounts))
	for i, a := range accounts {
		items[i] = a.Name
	}

	prompt := promptui.Select{
		Label: "Select Account",
		Items: items,
	}

	index, _, err := prompt.Run()
	if err != nil {
		return accountstore.Account{}, err
	}

	return accounts[index], nil
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
	logger.Debug("clearing accounts store")
	return accountstore.Clear()
}
