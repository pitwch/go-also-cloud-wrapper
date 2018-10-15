package alsocloud

import (
	"time"
)

type Options struct {
	Key           string
	APIPrefix     string
	LoginEndpoint string
	UserAgent     string
	Timeout       time.Duration
	VerifySSL     bool
	Batchsize     int
	Log           bool
}
