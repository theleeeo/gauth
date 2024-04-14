package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/theleeeo/thor/runner"
	"gopkg.in/yaml.v3"
)

func loadConfig() (*runner.Config, error) {
	content, err := os.ReadFile("./.thor.yml")
	if err != nil {
		log.Fatal(err)
	}

	var config runner.Config
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return &config, nil
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the server",
	RunE: func(cmd *cobra.Command, args []string) error {

		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		fmt.Println("Config:", cfg)
		fmt.Println("OauthCfg:", cfg.OAuthConfig)

		if err := runner.Run(cfg); err != nil {
			log.Println(err)
			return nil
		}

		return nil
	},
}
