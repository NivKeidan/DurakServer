package config



type settings struct {
	CorsOrigin			string
	CorsHeaders			string
}

func getSettings(env environment) *settings {
	switch env {
	case PROD:
		return &settings{
			CorsOrigin: "*",
			CorsHeaders: "Content-Type, ConnectionId",
		}
	case DEV:
		return &settings{
			CorsOrigin: "*",
			CorsHeaders: "Content-Type, ConnectionId",
		}
	case TEST:
		return &settings{
			CorsOrigin: "*",
			CorsHeaders: "Content-Type, ConnectionId",
		}
	case STAGING:
		return &settings{
			CorsOrigin: "*",
			CorsHeaders: "Content-Type, ConnectionId",
		}
	default:
		return nil
	}
}

func (this *settings) get(key string) string {
	switch key {
	case "CorsHeaders":
		return this.CorsHeaders
	case "CorsOrigin":
		return this.CorsOrigin
	default:
		return ""
	}
}