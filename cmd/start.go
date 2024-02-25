package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/theleeeo/thor/runner"
)

func init() {
	startCmd.Flags().StringP("addr", "a", "0.0.0.0:8080", "host:port to run the server on")
	cobra.CheckErr(viper.BindPFlag("addr", startCmd.Flags().Lookup("addr")))

	startCmd.Flags().String("secret-key", "", "secret key for JWT token signing")
	cobra.CheckErr(viper.BindPFlag("secret-key", startCmd.Flags().Lookup("secret-key")))

	startCmd.Flags().Duration("valid-duration", 5*time.Minute, "valid duration for JWT token")
	cobra.CheckErr(viper.BindPFlag("dvalid-duration", startCmd.Flags().Lookup("valid-duration")))
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the server",
	RunE: func(cmd *cobra.Command, args []string) error {

		cfg := &runner.Config{
			Addr:          viper.GetString("addr"),
			SecretKey:     viper.GetString("secret-key"),
			ValidDuration: viper.GetDuration("valid-duration"),
		}

		fmt.Println("Config:", cfg)

		r := runner.New(cfg)
		return r.Run()
	},
}
