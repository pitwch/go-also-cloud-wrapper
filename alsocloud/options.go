package alsocloud

import (
	"time"
)

//Options for Wrapper
//Can be changed on creating Client
type Options struct {
	APIPrefix     string
	LoginEndpoint string
	UserAgent     string
	Timeout       time.Duration
	VerifySSL     bool
	Batchsize     int
	Log           bool
}
