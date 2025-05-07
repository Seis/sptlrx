package cmd

import (
	"fmt"
	"strings"

	"github.com/raitonoberu/sptlrx/config"
	"github.com/raitonoberu/sptlrx/lyrics"
	"github.com/raitonoberu/sptlrx/pool"

	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"
	"github.com/spf13/cobra"
	"os"
)

var filePath string

var pipeCmd = &cobra.Command{
	Use:   "pipe",
	Short: "Start printing the current lines to stdout",

	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := loadConfig(cmd)
		if err != nil {
			return fmt.Errorf("couldn't load config: %w", err)
		}
		player, err := loadPlayer(conf)
		if err != nil {
			return fmt.Errorf("couldn't load player: %w", err)
		}
		provider, err := loadProvider(conf, player)
		if err != nil {
			return fmt.Errorf("couldn't load provider: %w", err)
		}

		ch := make(chan pool.Update)
		go pool.Listen(player, provider, conf, ch)

		for update := range ch {
			printUpdate(update, conf)
		}
		return nil
	},
}

var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "Start writing the current line to a file",

	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := loadConfig(cmd)
		if err != nil {
			return fmt.Errorf("couldn't load config: %w", err)
		}
		player, err := loadPlayer(conf)
		if err != nil {
			return fmt.Errorf("couldn't load player: %w", err)
		}
		provider, err := loadProvider(conf, player)
		if err != nil {
			return fmt.Errorf("couldn't load provider: %w", err)
		}

		ch := make(chan pool.Update)
		go pool.Listen(player, provider, conf, ch)

		for update := range ch {
			writeSingleLineToFile(update, conf, filePath)
		}
		return nil
	},
}

func writeToFile(content string, path string) {
	file, err := os.OpenFile(path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	_, err = file.WriteString(content + "\n")
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}
}

func printUpdate(update pool.Update, conf *config.Config) {
	if update.Err != nil {
		if !conf.IgnoreErrors {
			fmt.Println(update.Err.Error())
		}
		return
	}
	if update.Lines == nil || !lyrics.Timesynced(update.Lines) {
		fmt.Println("")
		return
	}
	line := update.Lines[update.Index].Words
	if conf.Pipe.Length == 0 {
		fmt.Println(line)
		return
	}
	switch conf.Pipe.Overflow {
	case "none":
		s := wrap.String(line, conf.Pipe.Length)
		fmt.Println(strings.Split(s, "\n")[0])
	case "word":
		s := wordwrap.String(line, conf.Pipe.Length)
		fmt.Println(strings.Split(s, "\n")[0])
	case "ellipsis":
		s := wrap.String(line, conf.Pipe.Length)
		lines := strings.Split(s, "\n")
		if len(lines) == 1 {
			fmt.Println(lines[0])
			return
		}
		s = wrap.String(lines[0], conf.Pipe.Length-3)
		fmt.Println(strings.Split(s, "\n")[0] + "...")
	}
}

func writeSingleLineToFile(update pool.Update, conf *config.Config, path string) {
	if update.Err != nil {
		if !conf.IgnoreErrors {
			fmt.Println(update.Err.Error())
		}
		return
	}
	if update.Lines == nil || !lyrics.Timesynced(update.Lines) {
		writeToFile("", path)
		return
	}
	line := update.Lines[update.Index].Words
	if conf.Pipe.Length == 0 {
		writeToFile(line, path)
		return
	}
	switch conf.Pipe.Overflow {
	case "none":
		s := wrap.String(line, conf.Pipe.Length)
		writeToFile(strings.Split(s, "\n")[0], path)
	case "word":
		s := wordwrap.String(line, conf.Pipe.Length)
		writeToFile(strings.Split(s, "\n")[0], path)
	case "ellipsis":
		s := wrap.String(line, conf.Pipe.Length)
		lines := strings.Split(s, "\n")
		if len(lines) == 1 {
			writeToFile(lines[0], path)
			return
		}
		s = wrap.String(lines[0], conf.Pipe.Length-3)
		writeToFile(strings.Split(s, "\n")[0]+"...", path)
	}
}
