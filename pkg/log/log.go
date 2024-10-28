package log

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

type Logger interface {
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Waiting(msg string) func(doneMsg string, success bool)
}

type ConsoleLogger struct {
	logger *log.Logger
}

func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{
		logger: log.New(os.Stdout, "", 0),
	}
}

func (c *ConsoleLogger) Info(msg string) {
	info := color.New(color.FgCyan).Sprintf("[INFO]: %s", msg)
	c.logger.Println(info)
}

func (c *ConsoleLogger) Warn(msg string) {
	warn := color.New(color.FgYellow).Sprintf("[WARN]️: %s", msg)
	c.logger.Println(warn)
}

func (c *ConsoleLogger) Error(msg string) {
	errMsg := color.New(color.FgRed).Sprintf("[ERROR]: %s", msg)
	c.logger.Println(errMsg)
}

func (c *ConsoleLogger) Waiting(msg string) func(doneMsg string, success bool) {
	c.Info(fmt.Sprintf("⏳ %s", msg))

	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Start()

	// Return a function to stop the spinner and mark completion
	return func(doneMsg string, success bool) {
		s.Stop()

		if success {
			c.logger.Println(color.New(color.FgGreen).Sprintf("[INFO]: ✅ %s", doneMsg))
		} else {
			c.logger.Println(color.New(color.FgRed).Sprintf("[INFO]: ❌ %s", doneMsg))
		}
	}
}
