package path

import (
	"os"
	"path/filepath"
	"runtime"
)

func GetConfigFile() (string, error) {
	const configDir, configFile = "kubedb", "config.yaml"
	var dir string
	switch runtime.GOOS {
	case "darwin":
		if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
			dir = filepath.Join(xdgConfigHome, configDir)
			break
		}
		fallthrough
	default:
		var err error
		dir, err = os.UserConfigDir()
		if err != nil {
			return "", err
		}

		dir = filepath.Join(dir, configDir)
	}
	return filepath.Join(dir, configFile), nil
}
