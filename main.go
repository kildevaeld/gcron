package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/kildevaeld/gcron/internal"
	flag "github.com/spf13/pflag"
)

var configFiles []string
var jsonOut bool
var versionFlag bool
var testFlag string

var VERSION string

func main() {

	if err := realMain(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}

}

func realMain() error {

	flag.StringSliceVarP(&configFiles, "file", "f", nil, "file")
	flag.BoolVar(&jsonOut, "json", false, "")
	flag.BoolVarP(&versionFlag, "version", "v", false, "")
	flag.StringVarP(&testFlag, "test", "t", "", "Test a job")
	flag.Parse()

	if jsonOut {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	if versionFlag {
		fmt.Printf("gcron version %s\n", VERSION)
		os.Exit(0)
	}

	c := internal.NewCron()

	if len(configFiles) == 0 {
		return errors.New("No cronfiles")
	}

	if err := loadFiles(c); err != nil {
		return err
	}

	if testFlag != "" {
		return testJob(testFlag, c)
	}

	c.Start()

	return listen(c)
}

func testJob(name string, c *internal.Cron) error {

	job := c.Get(name)
	if job == nil {
		return errors.New("no job with the name " + name)
	}

	

	return job.Run(os.Stdout, os.Stderr)
}

func printEvent(e internal.TaskEvent) {
	if start, ok := e.(*internal.StartEvent); ok {
		logrus.WithFields(logrus.Fields{
			"jobId": start.JobID(),
			"time":  start.Time(),
		}).Infof("Started %s", start.Name())

	} else if end, ok := e.(*internal.FinishEvent); ok {

		logger := logrus.WithFields(logrus.Fields{
			"jobId":    end.JobID(),
			"duration": end.Duration().String(),
		})

		if end.Error() != nil {
			logger.WithError(end.Error()).Errorf("Completed %s with error", end.Name())
		} else {
			logger.Infof("Completed %s", end.Name())
		}

	} else if loaded, ok := e.(*internal.LoadedEvent); ok {

		logger := logrus.WithFields(logrus.Fields{
			"jobId": loaded.JobID(),
		})
		logger.Infof("Loaded %s", loaded.Name())
	}
}

func loadFiles(c *internal.Cron) error {
	for _, file := range configFiles {
		if err := c.AddFile(file); err != nil {
			return err
		}
	}
	return nil
}

func listen(c *internal.Cron) error {
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
		case ev := <-c.Event():
			printEvent(ev)
		}
	}

}
