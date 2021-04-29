package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
	"gopkg.in/tucnak/telebot.v2"

	"github.com/huangnauh/tgscli/utils"
	"github.com/huangnauh/tgscli/version"
)

type Config struct {
	BotToken string
	telebot.StoredMessage
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
	return filepath.Join(configDir, "config.json")
}

func GetMetaPath() string {
	userDir := getDefaultCacheDir()
	return filepath.Join(userDir, "meta.json")
}

func GetMetaPinnedPath() string {
	userDir := getDefaultCacheDir()
	return filepath.Join(userDir, "meta_pinned.json")
}

func ReadConfig() (c Config, err error) {
	file, err := os.OpenFile(getDefaultConfigPath(), os.O_RDONLY, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, err
	}
	body, err := ioutil.ReadAll(file)
	if err != nil {
		return Config{}, err
	}

	err = json.Unmarshal(body, &c)
	if err != nil {
		return Config{}, err
	}
	return c, nil
}

func (c *Config) Write() error {
	fmt.Fprintln(outWriter, "Config write.")
	configPath := getDefaultConfigPath()

	file, err := os.OpenFile(configPath, os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	body, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return err
	}
	_, err = file.Write(body)
	return err
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
