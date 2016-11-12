package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/kildevaeld/gcron"
	flag "github.com/spf13/pflag"
)

var configFiles []string
var jsonOut bool

func main() {

	if err := realMain(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}

}

func realMain() error {

	flag.StringSliceVarP(&configFiles, "file", "f", nil, "file")
	flag.BoolVar(&jsonOut, "json", false, "")

	flag.Parse()

	if jsonOut {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	c := gcron.NewCron()

	if len(configFiles) == 0 {
		return errors.New("No cronfiles")
	}

	if err := loadFiles(c); err != nil {
		return err
	}

	c.Start()

	return listen(c)
}

func printEvent(e gcron.TaskEvent) {
	if start, ok := e.(*gcron.StartEvent); ok {
		logrus.WithFields(logrus.Fields{
			"jobID": start.JobID(),
			"time":  start.Time(),
		}).Infof("Started %s", start.Name())
	} else if end, ok := e.(*gcron.FinishEvent); ok {

		logger := logrus.WithFields(logrus.Fields{
			"jobId":    end.JobID(),
			"duration": end.Duration().String(),
		})

		if end.Error() != nil {
			logger.WithError(end.Error()).Errorf("Completed %s with error", end.Name())
		} else {
			logger.Infof("Completed %s", end.Name())
		}

	}
}

func loadFiles(c *gcron.Cron) error {
	for _, file := range configFiles {
		if err := c.AddFile(file); err != nil {
			return err
		}
	}
	return nil
}

func listen(c *gcron.Cron) error {
	sigs := make(chan os.Signal, 1)
	done := make(chan error, 1)
	defer close(sigs)
	defer close(done)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGUSR1)

	go func() {

		for {
			sig := <-sigs
			switch sig {
			case syscall.SIGUSR1:
				logrus.Infof("Reloading config")
				c.Clear()
				if err := loadFiles(c); err != nil {
					done <- err
					return
				}
				c.Start()
			case syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL:
				logrus.Infof("Existing...")
				done <- c.Stop()
				return
			}
		}
	}()

	for {
		select {
		case err := <-done:
			return err
		case err := <-c.Event():
			printEvent(err)
		}
	}

}
