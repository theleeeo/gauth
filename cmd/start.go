package cmd

import (
	"encoding/json"
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

		log.Println("Config:", prettyPrint(cfg))

		if err := runner.Run(cfg); err != nil {
			log.Println(err)
			return nil
		}

		return nil
	},
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "  ")
	return string(s)
}
