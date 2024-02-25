package runner

import "time"

type Config struct {
	Addr          string
	SecretKey     string
	ValidDuration time.Duration
}
