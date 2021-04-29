package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/huangnauh/tgscli/version"
)

func GetStrEnv(name string) string {
	return os.Getenv(fmt.Sprintf("%s_%s", strings.ToUpper(version.APP), strings.ToUpper(name)))
}

func GetDebugEnv() string {
	return GetStrEnv("debug")
}

func GetBoolEnv(name string) bool {
	e := GetStrEnv(name)
	return e != "" && strings.EqualFold(e, "false")
}

func GetScoopEnv() string {
	return os.Getenv("SCOOP")
}

func UserCacheDir() (string, error) {
	cacheDir := GetStrEnv("cache")
	if cacheDir == "" {
		return os.UserCacheDir()
	}
	return cacheDir, nil
}

func UserConfigDir() (string, error) {
	configDir := GetStrEnv("config")
	if configDir == "" {
		return os.UserConfigDir()
	}
	return configDir, nil
}
