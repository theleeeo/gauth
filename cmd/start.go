package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/theleeeo/thor/runner"
	"gopkg.in/yaml.v3"
)

func init() {
	startCmd.Flags().StringP("addr", "a", "0.0.0.0:8080", "host:port to run the server on")
	cobra.CheckErr(viper.BindPFlag("addr", startCmd.Flags().Lookup("addr")))

	startCmd.Flags().String("secret-key", "", "secret key for JWT token signing")
	cobra.CheckErr(viper.BindPFlag("secret-key", startCmd.Flags().Lookup("secret-key")))

	startCmd.Flags().Duration("valid-duration", 5*time.Minute, "valid duration for JWT token")
	cobra.CheckErr(viper.BindPFlag("dvalid-duration", startCmd.Flags().Lookup("valid-duration")))
}

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

		r := runner.New(cfg)
		return r.Run()
	},
}
