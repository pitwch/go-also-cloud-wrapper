package alsocloud

import (
	"time"
)

type Options struct {
	APIPrefix     string
	LoginEndpoint string
	UserAgent     string
	Timeout       time.Duration
	VerifySSL     bool
	Batchsize     int
	Log           bool
}
