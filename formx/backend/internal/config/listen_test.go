package config

import "testing"

func TestListenPortStaysOnServerPortWhenMorphPortIsSet(t *testing.T) {
	orig := osGetEnv
	t.Cleanup(func() { osGetEnv = orig })
	osGetEnv = func(key string) string {
		if key == "PORT" {
			return "9090"
		}
		return ""
	}

	cfg := Load()
	if cfg.ServerPort != "29909" {
		t.Fatalf("SERVER_PORT default = %q, PORT must not move the listener", cfg.ServerPort)
	}
	if cfg.FormsXSQLitePath != "./data/formsx.sqlite" || cfg.FormsXBadgerPath != "./data/formsx_badger" || cfg.UploadDir != "./uploads" {
		t.Fatalf("checkout defaults changed: %+v", cfg)
	}
}

func TestImageDataPaths(t *testing.T) {
	orig := osGetEnv
	t.Cleanup(func() { osGetEnv = orig })
	osGetEnv = func(key string) string {
		switch key {
		case "FORMSX_SQLITE_PATH":
			return "/data/formsx.sqlite"
		case "FORMSX_BADGER_PATH":
			return "/data/formsx_badger"
		case "UPLOAD_DIR":
			return "/data/uploads"
		case "SERVER_PORT":
			return "29909"
		default:
			return ""
		}
	}

	cfg := Load()
	if cfg.FormsXSQLitePath != "/data/formsx.sqlite" || cfg.FormsXBadgerPath != "/data/formsx_badger" || cfg.UploadDir != "/data/uploads" {
		t.Fatalf("image paths = %+v", cfg)
	}
	if cfg.ServerPort != "29909" {
		t.Fatalf("SERVER_PORT = %q", cfg.ServerPort)
	}
}
