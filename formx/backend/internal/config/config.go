package config

// Load reads configuration from environment with defaults.
func Load() *Config {
	return &Config{
		FormsXSQLitePath:  getEnv("FORMSX_SQLITE_PATH", "./data/formsx.sqlite"),
		FormsXBadgerPath:  getEnv("FORMSX_BADGER_PATH", "./data/formsx_badger"),
		ServerPort:        getEnv("SERVER_PORT", "29909"),
		UploadDir:         getEnv("UPLOAD_DIR", "./uploads"),
		SMTPHost:          getEnv("SMTP_HOST", ""),
		SMTPPort:          getEnv("SMTP_PORT", "587"),
		SMTPUser:          getEnv("SMTP_USER", ""),
		SMTPPassword:      getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:          getEnv("SMTP_FROM", ""),
		PublicFormBaseURL: getEnv("PUBLIC_FORM_BASE_URL", "http://localhost:19909"),
		UsersPanelBaseURL: getEnv("USERS_PANEL_BASE_URL", "http://127.0.0.1:9090"),
	}
}

type Config struct {
	FormsXSQLitePath  string
	FormsXBadgerPath  string
	ServerPort        string
	UploadDir         string
	SMTPHost          string
	SMTPPort          string
	SMTPUser          string
	SMTPPassword      string
	SMTPFrom          string
	PublicFormBaseURL string // used in broadcast emails for /f/{slug} links
	UsersPanelBaseURL string
}

func getEnv(key, defaultVal string) string {
	if v := osGetEnv(key); v != "" {
		return v
	}
	return defaultVal
}

// osGetEnv is overridable for tests
var osGetEnv = func(key string) string {
	// Will be set to os.Getenv in init or main
	return ""
}
