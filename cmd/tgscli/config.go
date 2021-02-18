package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/huangnauh/tgscli/pkg/utils"
	"github.com/huangnauh/tgscli/pkg/version"
)

type Config struct {
	BotToken   string
	ChatID     int64
	DatabaseID string
}

var cfg Config

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Handle configuration",
}

func init() {
	configCmd.AddCommand(configAddClusterCmd)
	rootCmd.AddCommand(configCmd)
}

func getDefaultConfigDir() string {
	configDir, err := utils.UserConfigDir()
	if err != nil {
		panic(err)
	}

	return filepath.Join(configDir, version.APP)
}

func getDefaultCacheDir() string {
	configDir, err := utils.UserCacheDir()
	if err != nil {
		panic(err)
	}

	return filepath.Join(configDir, version.APP)
}

func getDefaultConfigPath() string {
	configDir := getDefaultConfigDir()
	return filepath.Join(configDir, "config.yaml")
}

func GetDbPath() string {
	userDir := getDefaultCacheDir()
	return filepath.Join(userDir, "data.db")
}

func ReadConfig() (c Config, err error) {
	file, err := os.OpenFile(getDefaultConfigPath(), os.O_RDONLY, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, err
	}
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&c)
	if err != nil {
		return Config{}, err
	}
	return c, nil
}

func (c *Config) Write() error {
	configPath := getDefaultConfigPath()

	file, err := os.OpenFile(configPath, os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}

	encoder := yaml.NewEncoder(file)
	return encoder.Encode(&c)
}

var configAddClusterCmd = &cobra.Command{
	Use:   "set [BotToken] [ChatID]",
	Short: "set config",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg.BotToken = args[0]
		var err error
		cfg.ChatID, err = strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			errorExitf("Invalid ChatID: %v\n", err)
		}

		configDir := getDefaultConfigDir()
		_ = os.MkdirAll(configDir, 0644)
		userDir := getDefaultCacheDir()
		_ = os.MkdirAll(userDir, 0644)
		fmt.Fprintln(outWriter, "Config setted.")
	},
}
