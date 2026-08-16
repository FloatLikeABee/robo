package config

import "os"

func init() {
	osGetEnv = os.Getenv
}
