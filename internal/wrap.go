package internal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/kildevaeld/exec"
)

type count32 int32

func (c *count32) increment() int32 {
	return atomic.AddInt32((*int32)(c), 1)
}

func (c *count32) get() int32 {
	return atomic.LoadInt32((*int32)(c))
}

var counter count32

type wrapper struct {
	job     *CronJob
	c       chan<- TaskEvent
	logPath string
}

func (self *wrapper) Run() {

	name := self.job.config.Name
	if name == "" {
		name = fmt.Sprintf("task-%d", counter.increment())
	}

	now := time.Now()
	self.c <- &StartEvent{
		id:   self.job.id,
		time: now,
		name: name,
	}

	stdout, stderr, err := self.getStreams(self.job.config)
	if err != nil {
		self.c <- &ErrorEvent{}
	}

	err = self.job.Run(stdout, stderr)

	self.c <- &FinishEvent{
		id:       self.job.id,
		duration: time.Since(now),
		err:      err,
		name:     name,
	}

	outpath := stdout.Name()
	errpath := stderr.Name()

	if err != nil && self.job.config.OnError != nil {
		self.runHook(self.job.config.OnError, "onerror", name, outpath, errpath)
	} else if err == nil && self.job.config.OnComplete != nil {
		self.runHook(self.job.config.OnComplete, "onerror", name, outpath, errpath)
	}

	stdout.Close()
	stderr.Close()
	//os.Remove(outpath)
	//os.Remove(errpath)

}

func (self *wrapper) Stop() error {
	return self.job.Stop()
}

func (self *wrapper) getStreams(c CronJobConfig) (stdout *os.File, stderr *os.File, err error) {

	//stdoutPath := filepath.Join(self.logPath, fmt.Sprintf("%s-%s-%d.log", c.Name, self.job.id, time.Now().UnixNano()))
	//stderrPath := filepath.Join(self.logPath, fmt.Sprintf("%s-%s-%d.err", c.Name, self.job.id, time.Now().UnixNano()))

	stdoutPath := filepath.Join(self.logPath, fmt.Sprintf("gcron-task-%s-%s.log", c.Name, self.job.id))
	stderrPath := filepath.Join(self.logPath, fmt.Sprintf("gcron-task-%s-%s.err", c.Name, self.job.id))

	if stdout, err = os.Create(stdoutPath); err != nil {
		// error
		return nil, nil, err
	}

	if stderr, err = os.Create(stderrPath); err != nil {
		stdout.Close()
		// error
		return nil, nil, err
	}

	return
}

func (self *wrapper) runHook(hook *JobConfig, hookname, name, stdout, stderr string) {

	job, err := self.job.getConfig(*hook, nil, nil)
	if err != nil {
		self.c <- &ErrorEvent{}
		return
	}

	job.Args = append(job.Args, stdout, stderr)

	now := time.Now()
	cmd := exec.New(*job)
	self.c <- &StartEvent{
		id:   self.job.id,
		name: fmt.Sprintf("%s %s hook", name, hookname),
		time: now,
	}

	err = cmd.Start(context.Background())

	self.c <- &FinishEvent{
		id:       self.job.id,
		name:     fmt.Sprintf("%s %s hook", name, hookname),
		duration: time.Since(now),
		err:      err,
	}

}
