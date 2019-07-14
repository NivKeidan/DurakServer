package config

import "sync"

// Singleton

type Configuration struct {
	env      environment
	settings *settings
}

var instance *Configuration
var once sync.Once

func GetConfiguration(envStr string) *Configuration {

	once.Do(func() {
		var env environment
		switch envStr {
		case "PROD":
			env = PROD
		case "DEV":
			env = DEV
		case "TEST":
			env = TEST
		case "STAGING":
			env = STAGING
		default:
			env = environment("")
		}

		if env != environment("") {
			instance = &Configuration{
				env:      env,
				settings: getSettings(env),
			}
		} else {
			instance = nil
		}
	})
	return instance
}

func (this *Configuration) GetString(key string) string {
	// If no configuration was loaded, use default of DEV
	// TODO Change default environment to prod?
	
	defaultEnv := "DEV"

	if instance == nil {
		GetConfiguration(defaultEnv)
	}
	return instance.settings.getString(key)
}

func (this *Configuration) GetInt(key string) int {
	// If no configuration was loaded, use default of DEV
	// TODO Change default environment to prod?

	defaultEnv := "DEV"

	if instance == nil {
		GetConfiguration(defaultEnv)
	}
	return instance.settings.getInt(key)
}
