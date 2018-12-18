package path

import (
	"os"
	"path/filepath"
	"runtime"

	homedir "github.com/mitchellh/go-homedir"
)

const (
	darwin  = "darwin"
	freebsd = "freebsd"
	windows = "windows"
)

func windowsProgramDataPath() string {
	programDataPath := os.Getenv("ALLUSERSPROFILE")
	if programDataPath == "" {
		programDataPath = "C:\\ProgramData"
	}
	return programDataPath
}

// SystemConfigDir returns the path to the sensu config directory based on the runtime OS
func SystemConfigDir() string {
	switch runtime.GOOS {
	case windows:
		return filepath.Join(windowsProgramDataPath(), "sensu", "config")
	case freebsd:
		return filepath.Join("/usr/local/etc/sensu")
	default:
		return filepath.Join("/etc/sensu")
	}
}

// SystemCacheDir returns the path to the sensu cache directory based on the runtime OS
func SystemCacheDir(exeName string) string {
	switch runtime.GOOS {
	case windows:
		return filepath.Join(windowsProgramDataPath(), "sensu", "cache", exeName)
	default:
		return filepath.Join("/var/cache/sensu", exeName)
	}
}

// SystemDataDir returns the path to the data (state) directory based on the runtime OS
func SystemDataDir(exeName string) string {
	switch runtime.GOOS {
	case windows:
		return filepath.Join(windowsProgramDataPath(), "sensu", "data", exeName)
	default:
		return filepath.Join("/var/lib/sensu", exeName)
	}
}

// SystemPidDir returns the path to the pid directory based on the runtime OS
func SystemPidDir() string {
	switch runtime.GOOS {
	case windows:
		// TODO (JK): do we need a pid dir on windows?
		return filepath.Join(windowsProgramDataPath(), "sensu", "run")
	default:
		return filepath.Join("/var/run/sensu")
	}
}

// SystemLogDir returns the path to the log directory based on the runtime OS
func SystemLogDir() string {
	switch runtime.GOOS {
	case windows:
		return filepath.Join(windowsProgramDataPath(), "sensu", "log")
	default:
		return filepath.Join("/var/log/sensu")
	}
}

// UserConfigDir returns the path to the current user's config directory based on the runtime OS
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

// UserCacheDir returns the path to the current user's cache directory based on the runtime OS
func UserCacheDir(exeName string) string {
	switch runtime.GOOS {
	case windows:
		localAppDataPath := os.Getenv("LOCALAPP")
		if localAppDataPath == "" {
			h, _ := homedir.Dir()
			localAppDataPath = filepath.Join(h, "AppData", "Local")
		}
		return filepath.Join(localAppDataPath, "sensu", exeName)
	case darwin:
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
