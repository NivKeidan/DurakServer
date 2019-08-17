package config



type settings struct {
	CorsOrigin			string
	CorsHeaders			string
	clientIdLetters		string
	clientIdLength		int
}

func getSettings(env environment) *settings {

	// Relevant to all environments
	s := &settings{
		CorsHeaders: "Content-Type, ConnectionId",
		clientIdLetters: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		clientIdLength: 10,
	}

	// Unique varlues er environment
	switch env {
	case PROD:
		s.CorsOrigin = "*"
		return s
	case DEV:
		s.CorsOrigin = "*"
		return s
	case TEST:
		s.CorsOrigin = "*"
		return s
	case STAGING:
		s.CorsOrigin = "*"
		return s
	default:
		return nil
	}
}

func (this *settings) getString(key string) string {
	switch key {
	case "CorsHeaders":
		return this.CorsHeaders
	case "CorsOrigin":
		return this.CorsOrigin
	case "ClientIdLetters":
		return this.clientIdLetters
	default:
		return ""
	}
}

func (this *settings) getInt(key string) int {
	switch key {
	case "ClientIdLength":
		return this.clientIdLength
	case "AliveTTL":
		return 5
	default:
		return 0
	}
}