package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/terzigolu/josepshbrain-go/internal/config"
	"github.com/urfave/cli/v2"
)

func NewSetupCommand() *cli.Command {
	return &cli.Command{
		Name:  "setup",
		Usage: "Configure the CLI with your API key",
		Action: func(c *cli.Context) error {
			reader := bufio.NewReader(os.Stdin)

			fmt.Print("Enter your API Key: ")
			apiKey, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("could not read API key: %w", err)
			}
			apiKey = strings.TrimSpace(apiKey)

			cfg, err := config.LoadConfig()
			if err != nil {
				// If config doesn't exist, a new one will be created.
				// We can ignore the error for now and proceed with a new config object.
				cfg = &config.Config{}
			}

			cfg.APIKey = apiKey

			err = config.SaveConfig(cfg)
			if err != nil {
				return fmt.Errorf("could not save config: %w", err)
			}

			fmt.Println("✅ Configuration saved successfully!")
			return nil
		},
	}
} 