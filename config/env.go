package config

type environment string

const (
	PROD  = environment("Production")
	DEV = environment("Development")
	TEST = environment("Testing")
	STAGING = environment("Staging")
)
