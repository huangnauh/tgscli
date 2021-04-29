package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/dustin/go-humanize"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gopkg.in/tucnak/telebot.v2"
)

func init() {
	rootCmd.AddCommand(putCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(shareCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "verbose")
	rootCmd.PersistentFlags().BoolVarP(&forceFlag, "force", "f", false, "force")
	rootCmd.PersistentFlags().BoolVar(&savePinnedFlag, "save-pinned", false, "save pinned meta")
}

var (
	bot            *telebot.Bot
	chat           *telebot.Chat
	forceFlag      bool
	verboseFlag    bool
	savePinnedFlag bool
	messages       map[string]*Message
)

type Message struct {
	telebot.File
	telebot.StoredMessage
	Time time.Time `json:"time"`
}

func (m *Message) GetEditable() telebot.Editable {
	return m.StoredMessage
}

func (m *Message) Marshal() []byte {
	body, err := json.Marshal(m)
	if err != nil {
		errorExitf("msg Info %+v: %v\n", m, err)
	}
	return body
}

func postrun(cmd *cobra.Command, args []string) {
}

func SaveMeta(metaPath string) {
	file, err := os.OpenFile(metaPath, os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		errorExitf("Open meta File %s: %s\n", metaPath, err)
	}
	body, err := json.Marshal(messages)
	if err != nil {
		errorExitf("Marshal meta File %s: %s\n", metaPath, err)
	}
	_, err = file.Write(body)
	if err != nil {
		errorExitf("save meta File %s: %s\n", metaPath, err)
	}
}

func GetMeta(metaPath string, metaInfo os.FileInfo) {
	if verboseFlag {
		fmt.Fprintf(outWriter, "local edited %s\n", metaInfo.ModTime())
	}
	metaFile, err := os.OpenFile(metaPath, os.O_RDONLY, 0644)
	if err != nil {
		errorExitf("Open meta File %s: %s\n", metaPath, err)
	}
	body, err := ioutil.ReadAll(metaFile)
	if err != nil {
		metaFile.Close()
		errorExitf("Read meta File %s: %s\n", metaPath, err)
	}
	err = json.Unmarshal(body, &messages)
	if err != nil {
		metaFile.Close()
		errorExitf("Unmarshal metadata %s: %s\n", metaPath, err)
	}
	metaFile.Close()
}

func prerun(_ *cobra.Command, _ []string) {
	var err error
	cfg, err = ReadConfig()
	if err != nil {
		errorExitf("Read Config: %v\n", err)
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
	prerunMessage()
}

func prerunMessage() {
	messages = make(map[string]*Message)
	metaPath := GetMetaPath()
	metaInfo, err := os.Stat(metaPath)
	if err != nil && os.IsNotExist(err) {
	} else if err != nil {
		errorExitf("Stat meta path: %s\n", err)
	} else {
		GetMeta(metaPath, metaInfo)
	}

	fmt.Fprintf(outWriter, "%d files in local\n", len(messages))
	if chat.PinnedMessage == nil {
		return
	}

	var editTime time.Time
	caption := chat.PinnedMessage.Document.Caption
	if strings.HasPrefix(caption, "meta-") && strings.HasSuffix(caption, ".json") {
		timestamp, _ := strconv.Atoi(caption[len("meta-") : len(caption)-len(".json")])
		if timestamp > 0 {
			editTime = time.Unix(int64(timestamp), 0)
		}
	} else {
		editTime = chat.PinnedMessage.LastEdited()
	}

	if verboseFlag {
		fmt.Fprintf(outWriter, "  pin edited %s\n", editTime)
	}
	if !forceFlag {
		if metaInfo != nil && !editTime.After(metaInfo.ModTime()) {
			return
		}
	}

	getPinnedMessages()
	if messageID, _ := chat.PinnedMessage.MessageSig(); messageID != cfg.MessageID {
		_ = cfg.Write()
	}
	SaveMeta(metaPath)
}

func getPinnedMessages() {
	buf := &bytes.Buffer{}
	err := GetFileByID(chat.PinnedMessage.Document.File.FileID,
		int64(chat.PinnedMessage.Document.File.FileSize), buf)
	if err != nil {
		errorExitf("pinned file: %s\n", err)
	}
	if savePinnedFlag {
		SaveMeta(GetMetaPinnedPath())
	}
	pinnedMessages := make(map[string]*Message)
	err = json.Unmarshal(buf.Bytes(), &pinnedMessages)
	if err != nil {
		errorExitf("pinned file: %s\n", err)
	}
	fmt.Fprintf(outWriter, "%d files in pinned\n", len(pinnedMessages))
	for name, pinnedMessage := range pinnedMessages {
		message, ok := messages[name]
		if !ok || message.FileID == pinnedMessage.FileID {
			messages[name] = pinnedMessage
		} else {
			fmt.Fprintf(outWriter, "message %s conflict with pinned, try new name\n", name)
			newName := getName(name, messages, pinnedMessages)
			messages[newName] = pinnedMessage
		}
	}
	fmt.Fprintf(outWriter, "got %d files in total\n", len(messages))
}

func GetURLByName(filename string) (string, error) {
	msg, ok := messages[filename]
	if !ok {
		return "", os.ErrNotExist
	}
	uri, err := bot.FileURLByID(msg.FileID)
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") {
			return "", os.ErrNotExist
		}
		return "", fmt.Errorf("get file url %s: %s", filename, err)
	}
	return uri, nil
}

func SaveFileByName(filename, output string) error {
	msg, ok := messages[filename]
	if !ok {
		return os.ErrNotExist
	}
	rawFile, err := os.OpenFile(output, os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("download file %s: %s", filename, err)
	}
	defer rawFile.Close()

	return GetFileByID(msg.FileID, int64(msg.FileSize), rawFile)
}

func GetFileByID(fileID string, size int64, writer io.Writer) error {
	uri, err := bot.FileURLByID(fileID)
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") {
			return os.ErrNotExist
		}
		return fmt.Errorf("get file uri %s: %s", fileID, err)
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		uri,
		nil,
	)
	if err != nil {
		return fmt.Errorf("get file req %s: %v", fileID, err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("get file %s: %v", fileID, err)
	}
	defer res.Body.Close()
	if res.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("get file response %s: %d", fileID, res.StatusCode)
	}

	var src io.Reader
	if size > 0 {
		bar := pb.Full.Start64(size)
		barReader := bar.NewProxyReader(res.Body)
		defer bar.Finish()
		src = barReader
	} else {
		src = res.Body
	}

	_, err = io.Copy(writer, src)
	if err != nil {
		return fmt.Errorf("download file %s: %s", fileID, err)
	}
	return nil
}

func UploadFileByName(filename string) *telebot.Message {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		errorExitf("Open File: %v\n", err)
	}
	info, err := file.Stat()
	if err != nil {
		errorExitf("Stat File: %v\n", err)
	}

	bar := pb.Full.Start64(info.Size())
	barReader := bar.NewProxyReader(file)
	defer bar.Finish()

	defer file.Close()
	name := filepath.Base(filename)
	m, err := bot.Send(chat, &telebot.Document{
		File:     telebot.FromReader(barReader),
		FileName: name,
	})
	if err != nil {
		errorExitf("Upload failed: %v\n", err)
	}
	return m
}

var getCmd = &cobra.Command{
	Use:   "get [file] [output]",
	Short: "get file",
	Args:  cobra.RangeArgs(1, 2),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		prerun(cmd, args)
	},
	PersistentPostRun: postrun,
	Run: func(cmd *cobra.Command, args []string) {
		outputFile := args[0]
		if len(args) == 2 {
			outputFile = args[1]
		}
		err := SaveFileByName(args[0], outputFile)
		if err != nil {
			errorExitf("Download: %v\n", err)
		}
		fmt.Fprintf(outWriter, "Download File to %s\n", outputFile)
	},
}

var shareCmd = &cobra.Command{
	Use:   "share [filename]",
	Short: "generate url for download access",
	Args:  cobra.ExactArgs(1),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		prerun(cmd, args)
	},
	PersistentPostRun: postrun,
	Run: func(cmd *cobra.Command, args []string) {
		uri, err := GetURLByName(args[0])
		if err != nil {
			errorExitf("Get url: %v\n", err)
		}
		fmt.Fprintf(outWriter, "Share :%s\n", uri)
	},
}

func formatTable(table *tablewriter.Table, name string, msg *Message) {
	table.Append([]string{
		msg.Time.Format(time.RFC1123),
		humanize.Bytes(uint64(msg.FileSize)),
		name,
	})
}

func newTable() *tablewriter.Table {
	table := tablewriter.NewWriter(outWriter)
	table.SetBorder(false)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader([]string{"Time", "Length", "Name"})
	return table
}

var listCmd = &cobra.Command{
	Use:   "list [contains]",
	Short: "list files containing relevant characters",
	Args:  cobra.MaximumNArgs(1),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		prerun(cmd, args)
	},
	PersistentPostRun: postrun,
	Run: func(cmd *cobra.Command, args []string) {
		table := newTable()
		contains := ""
		if len(args) == 1 {
			contains = args[0]
		}
		for name, message := range messages {
			if strings.Contains(name, contains) {
				formatTable(table, name, message)
			}
		}
		table.Render()
	},
}

func getName(filename string, messages ...map[string]*Message) string {
	name := filepath.Base(filename)
	for count := 1; ; count++ {
		found := false
		for _, msgMap := range messages {
			_, ok := msgMap[name]
			if ok {
				found = true
				break
			}
		}
		if !found {
			return name
		}
		name = fmt.Sprintf("%s#%d", name, count)
	}
}

func setName(messages map[string]*Message, filename string, msg *Message) error {
	name := getName(filename, messages)
	body, err := json.Marshal(messages)
	if err != nil {
		return err
	}
	metaPath := GetMetaPath()
	metaFile, err := os.OpenFile(metaPath, os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer metaFile.Close()
	_, err = metaFile.Write(body)
	messages[name] = msg
	return err
}

var putCmd = &cobra.Command{
	Use:   "put [file]",
	Short: "put file",
	Args:  cobra.ExactArgs(1),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		prerun(cmd, args)
	},
	PersistentPostRun: postrun,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		startTime := time.Now()
		m := UploadFileByName(args[0])
		fmt.Fprintf(outWriter, "Upload time:%s\n", time.Since(startTime))
		messageID, chatID := m.MessageSig()
		msg := &Message{
			m.Document.File,
			telebot.StoredMessage{
				MessageID: messageID,
				ChatID:    chatID,
			},
			time.Now(),
		}
		err = setName(messages, args[0], msg)
		if err != nil {
			errorExitf("set name to message: %v\n", err)
		}

		body, err := json.Marshal(messages)
		if err != nil {
			errorExitf("marshal messages: %v\n", err)
		}
		doc := &telebot.Document{
			File:     telebot.FromReader(bytes.NewReader(body)),
			Caption:  fmt.Sprintf("meta-%d.json", time.Now().Unix()),
			FileName: "meta.json",
		}
		if cfg.MessageID == "" {
			m, err = bot.Send(chat, doc)
			if err != nil {
				errorExitf("Upload failed: %v\n", err)
			}
			cfg.MessageID, _ = m.MessageSig()
			err = bot.Pin(cfg.StoredMessage)
			if err != nil {
				errorExitf("pin error: %v\n", err)
			}
			err = cfg.Write()
			if err != nil {
				errorExitf("write config: %v\n", err)
			}
		} else {
			m, err = bot.Edit(cfg.StoredMessage, doc)
			if err != nil {
				errorExitf("Upload failed: %v\n", err)
			}
			messageID, _ := m.MessageSig()
			if err != nil {
				errorExitf("write config: %v\n", err)
			}
			if messageID != cfg.MessageID {
				cfg.MessageID = messageID
				err = cfg.Write()
				if err != nil {
					errorExitf("write config: %v\n", err)
				}
			}
		}
		SaveMeta(GetMetaPath())
	},
}
