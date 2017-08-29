package path

import (
	"os"
	"path/filepath"
	"runtime"

	homedir "github.com/mitchellh/go-homedir"
)

const (
	windows = "windows"
)

func windowsProgramDataPath() string {
	programDataPath := os.Getenv("ALLUSERSPROFILE")
	if programDataPath == "" {
		programDataPath = "C:\\ProgramData"
	}
	return programDataPath
}

// SystemConfigDir ...
func SystemConfigDir() string {
	switch runtime.GOOS {
	case windows:
		return filepath.Join(windowsProgramDataPath(), "sensu", "config")
	case "freebsd":
		return filepath.Join("/usr/local/etc/sensu")
	default:
		return filepath.Join("/etc/sensu")
	}
}

// SystemCacheDir ...
func SystemCacheDir(exeName string) string {
	switch runtime.GOOS {
	case windows:
		return filepath.Join(windowsProgramDataPath(), "sensu", "cache", exeName)
	default:
		return filepath.Join("/var/cache/sensu", exeName)
	}
}

// SystemDataDir ...
func SystemDataDir() string {
	switch runtime.GOOS {
	case windows:
		return filepath.Join(windowsProgramDataPath(), "sensu", "data")
	default:
		return filepath.Join("/var/lib/sensu")
	}
}

// SystemPidDir ...
func SystemPidDir() string {
	switch runtime.GOOS {
	case windows:
		return filepath.Join(windowsProgramDataPath(), "sensu", "run")
	default:
		return filepath.Join("/var/run/sensu")
	}
}

// SystemLogDir ...
func SystemLogDir() string {
	switch runtime.GOOS {
	case windows:
		return filepath.Join(windowsProgramDataPath(), "sensu", "log")
	default:
		return filepath.Join("/var/log/sensu")
	}
}

// UserConfigDir ...
func UserConfigDir(exeName string) string {
	switch runtime.GOOS {
	case windows:
		appDataPath := os.Getenv("APPDATA")
		if appDataPath == "" {
			h, _ := homedir.Dir()
			appDataPath = filepath.Join(h, "AppData", "Roaming")
		}
		return filepath.Join(appDataPath, "sensu", exeName)
	default:
		xdgConfigPath := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfigPath == "" {
			h, _ := homedir.Dir()
			xdgConfigPath = filepath.Join(h, ".config")
		}
		return filepath.Join(xdgConfigPath, "sensu", exeName)
	}
}

// UserCacheDir ...
func UserCacheDir(exeName string) string {
	switch runtime.GOOS {
	case windows:
		localAppDataPath := os.Getenv("LOCALAPP")
		if localAppDataPath == "" {
			h, _ := homedir.Dir()
			localAppDataPath = filepath.Join(h, "AppData", "Local")
		}
		return filepath.Join(localAppDataPath, "sensu", exeName)
	case "darwin":
		h, _ := homedir.Dir()
		return filepath.Join(h, "Library", "Caches", "sensu", exeName)
	default:
		xdgCachePath := os.Getenv("XDG_CACHE_HOME")
		if xdgCachePath == "" {
			h, _ := homedir.Dir()
			xdgCachePath = filepath.Join(h, ".cache")
		}
		return filepath.Join(xdgCachePath, "sensu", exeName)
	}
}
