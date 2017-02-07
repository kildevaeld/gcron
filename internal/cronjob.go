package internal

import (
	"context"
	"errors"

	"github.com/kildevaeld/exec"
	system "github.com/kildevaeld/go-system"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type JobConfig struct {
	Workdir     string        `json:"workdir"`
	Command     string        `json:"command,omitempty"`
	Script      string        `json:"script,omitempty"`
	Env         exec.Environ  `json:"env"`
	SysEnv      bool          `json:"sysenv"`
	Interpreter []string      `json:"interpreter"`
	User        string        `json:"user"`
	Stdout      string        `json:"stdout"`
	Stderr      string        `json:"stderr"`
	Timeout     time.Duration `json:"duration"`
}

type CronJobConfig struct {
	JobConfig  `yaml:",inline"`
	Name       string     `json:"name"`
	Cron       string     `json:"cron"`
	Parallel   bool       `json:"parallel"`
	OnError    *JobConfig `json:"onerror,omitempty"`
	OnComplete *JobConfig `json:"oncomplete,omitempty"`
}

type CronJob struct {
	id      string
	config  CronJobConfig
	running bool
	lock    sync.Mutex
	cmd     *exec.Executor
	c       chan<- error
}

func (self *CronJob) run(stdout io.Writer, stderr io.Writer) error {

	if err := self.initExecutor(stdout, stderr); err != nil {
		return err
	}

	ctx := context.Background()

	if self.config.Timeout > 0 {
		ctx, _ = context.WithTimeout(ctx, self.config.Timeout)
	}

	err := self.cmd.Start(ctx)

	self.lock.Lock()

	self.running = false
	self.cmd = nil

	self.lock.Unlock()

	return err
}

func (self *CronJob) Run(stdout, stderr io.Writer) error {

	if self.IsRunning() && !self.config.Parallel {
		return errors.New("already running")
	}

	result := self.run(stdout, stderr)

	if result != nil && len(self.config.Interpreter) > 0 && self.config.Interpreter[0] == "notto" {
		stderr.Write([]byte(result.Error()))
	}

	return result
}

func (self *CronJob) initExecutor(stdout io.Writer, stderr io.Writer) error {

	return self.withLock(func() error {
		conf, e := self.getConfig(self.config.JobConfig, stdout, stderr)
		if e != nil {
			return e
		}

		self.cmd = exec.New(*conf)
		self.running = true
		return nil
	})

}

func (self *CronJob) getConfig(c JobConfig, stdout io.Writer, stderr io.Writer) (*exec.Config, error) {

	var cmds []string
	if c.Command != "" {
		cmds = strings.Split(c.Command, " ")
	}

	env := c.Env
	if c.SysEnv {
		env = exec.MergeEnviron(exec.Environ(os.Environ()), c.Env)
	}

	config := exec.Config{
		Cmd:         cmds,
		Script:      c.Script,
		WorkDir:     c.Workdir,
		Env:         env,
		Interpreter: c.Interpreter,
		Stdout:      stdout,
		Stderr:      stderr,
	}

	if c.User != "" {
		user, err := system.GetUser(c.User)
		if err != nil {
			return nil, err
		}
		config.User = user
	}

	return &config, nil
}

func (self *CronJob) withLock(fn func() error) error {
	self.lock.Lock()
	err := fn()
	self.lock.Unlock()
	return err
}

func (self *CronJob) IsRunning() bool {
	r := false
	self.withLock(func() error {
		r = self.running
		return nil
	})
	return r
}

func (self *CronJob) Stop() error {

	return self.withLock(func() error {
		if !self.running {
			return nil
		}

		if self.cmd == nil {
			return errors.New("Invalid state")
		}

		err := self.cmd.Stop()

		self.running = false
		self.cmd = nil

		return err
	})

}

func NewJob(config CronJobConfig, id string) *CronJob {
	return &CronJob{
		id:     id,
		config: config,
	}
}
