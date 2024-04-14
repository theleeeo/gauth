package authorizer

import "time"

type Config struct {
	AppUrl        string
	PrivateKey    []byte
	PublicKey     []byte
	ValidDuration time.Duration
}
