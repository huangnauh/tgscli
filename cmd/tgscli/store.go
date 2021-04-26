package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
	"gopkg.in/tucnak/telebot.v2"
)

func init() {
	rootCmd.AddCommand(putCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(listCmd)
}

var (
	db   *bolt.DB
	bot  *telebot.Bot
	chat *telebot.Chat
)

func postrun(cmd *cobra.Command, args []string) {
	db.Close()
}

func prerun(cmd *cobra.Command, args []string, remoteMetadata bool) {
	var err error
	cfg, err = ReadConfig()
	if err != nil {
		errorExitf("Read Config: %v\n", err)
	}

	dbPath := GetDbPath()
	_, err = os.Stat(dbPath)
	if err != nil && os.IsNotExist(err) {
		if remoteMetadata && cfg.DatabaseID != "" {
			SaveFileByID(cfg.DatabaseID, dbPath)
		}
	} else if err != nil {
		errorExitf("Open Database: %s\n", err)
	}
	db, err = bolt.Open(dbPath, 0666, nil)
	if err != nil {
		errorExitf("Open Database: %v\n", err)
	}

	bot, err = telebot.NewBot(telebot.Settings{
		Token: cfg.BotToken,
	})
	if err != nil {
		errorExitf("Create Telegram Bot: %v", err)
	}

	chat, err = bot.ChatByID(
		strconv.FormatInt(cfg.ChatID, 10),
	)
	if err != nil {
		errorExitf("Get Telegram Chat: %v", err)
	}
}

func SaveFileByName(filename, output string) {
	var v []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(strconv.FormatInt(cfg.ChatID, 10)))
		if b == nil {
			return nil
		}
		v = b.Get([]byte(filename))
		return nil
	})
	if v == nil {
		fmt.Fprintf(outWriter, "Empty\n")
		return
	}
	file := NewFile(v)
	err := SaveFileByID(file.FileID, output)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(outWriter, "Not Found\n")
		} else {
			errorExit(err)
		}
	}
	fmt.Fprintf(outWriter, "Download File to %s\n", output)
}

func SaveFileByID(fileID, output string) error {
	uri, err := bot.FileURLByID(fileID)
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") {
			return os.ErrNotExist
		}
		return fmt.Errorf("Get File Uri %s: %v\n", fileID, err)
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		uri,
		nil,
	)
	if err != nil {
		return fmt.Errorf("Get File Req %s: %v\n", fileID, err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("Get File %s: %v\n", fileID, err)
	}
	defer res.Body.Close()
	if res.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("Get File Response %s: %d\n", fileID, res.StatusCode)
	}

	rawFile, err := os.OpenFile(output, os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("Download File %s: %s\n", fileID, err)
	}
	defer rawFile.Close()
	_, err = io.Copy(rawFile, res.Body)
	if err != nil {
		return fmt.Errorf("Download File %s: %s\n", fileID, err)
	}
	return nil
}

func UploadFileByName(filename string) *telebot.Message {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		errorExitf("Upload File: %v\n", err)
	}

	defer file.Close()
	name := filepath.Base(filename)
	m, err := bot.Send(chat, &telebot.Document{
		File:     telebot.FromReader(file),
		FileName: name,
	})
	if err != nil {
		errorExitf("Upload failed: %v\n", err)
	}
	return m
}

type File struct {
	telebot.File
	Time time.Time `json:"time"`
}

func NewFile(v []byte) *File {
	file := &File{}
	err := json.Unmarshal(v, file)
	if err != nil {
		errorExitf("File Info %s: %v\n", v, err)
	}
	return file
}

func format(table *tablewriter.Table, k, v []byte) {
	file := NewFile(v)
	table.Append([]string{
		file.Time.Format(time.RFC1123),
		fmt.Sprintf("%d", file.FileSize),
		string(k)})
}

var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "upload metadata",
	Args:  cobra.NoArgs,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		prerun(cmd, args, false)
	},
	PersistentPostRun: postrun,
	Run: func(cmd *cobra.Command, args []string) {
		dbPath := GetDbPath()
		_, err := os.Stat(dbPath)
		if err != nil && os.IsNotExist(err) {
			fmt.Fprintf(outWriter, "Not Found\n")
			return
		}
		db, err = bolt.Open(dbPath, 0666, nil)
		if err != nil {
			errorExitf("Open Database: %v\n", err)
		}
		m := UploadFileByName(dbPath)
		cfg.DatabaseID = m.Document.File.FileID
		cfg.Write()
		fmt.Fprintf(outWriter, "Upload metadata to :%s\n", cfg.DatabaseID)
	},
}

var getCmd = &cobra.Command{
	Use:   "get [file] [output]",
	Short: "get file",
	Args:  cobra.RangeArgs(1, 2),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		prerun(cmd, args, true)
	},
	PersistentPostRun: postrun,
	Run: func(cmd *cobra.Command, args []string) {
		outputFile := args[0]
		if len(args) == 2 {
			outputFile = args[1]
		}
		SaveFileByName(args[0], outputFile)
	},
}

var listCmd = &cobra.Command{
	Use:   "list [prefix]",
	Short: "list prefix file",
	Args:  cobra.MaximumNArgs(1),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		prerun(cmd, args, true)
	},
	PersistentPostRun: postrun,
	Run: func(cmd *cobra.Command, args []string) {
		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(strconv.FormatInt(cfg.ChatID, 10)))
			if b == nil {
				fmt.Fprintf(outWriter, "Empty\n")
				return nil
			}
			c := b.Cursor()

			table := tablewriter.NewWriter(outWriter)
			table.SetBorder(false)
			table.SetCenterSeparator("")
			table.SetColumnSeparator("")
			table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.SetHeader([]string{"LastWriteTime", "Length", "Name"})

			if len(args) > 0 {
				prefix := []byte(args[0])
				for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
					format(table, k, v)
				}
			} else {
				for k, v := c.First(); k != nil; k, v = c.Next() {
					format(table, k, v)
				}
			}
			table.Render()
			return nil
		})
	},
}

var putCmd = &cobra.Command{
	Use:   "put [file]",
	Short: "put file",
	Args:  cobra.ExactArgs(1),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		prerun(cmd, args, true)
	},
	PersistentPostRun: postrun,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		startTime := time.Now()
		m := UploadFileByName(args[0])
		fmt.Fprintf(outWriter, "Upload time:%s\n", time.Since(startTime))
		body, _ := json.Marshal(File{
			m.Document.File,
			time.Now(),
		})

		err = db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte(strconv.FormatInt(cfg.ChatID, 10)))
			if err != nil {
				errorExitf("Put Cache: %v\n", err)
			}
			err = b.Put([]byte(filepath.Base(args[0])), body)
			if err != nil {
				errorExitf("Put Cache: %v\n", err)
			}
			return nil
		})
		if err != nil {
			errorExitf("Put Cache DB: %v\n", err)
		}
	},
}
