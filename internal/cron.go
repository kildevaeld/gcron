package internal

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"sync"

	yaml "gopkg.in/yaml.v2"

	"io/ioutil"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/pborman/uuid"
	"github.com/robfig/cron"
)

type Cron struct {
	cron *cron.Cron
	jobs []*wrapper
	lock sync.RWMutex
	tmp  string
	c    chan TaskEvent
}

func (self *Cron) Event() <-chan TaskEvent {
	return self.c
}

func (self *Cron) Start() {
	self.cron.Start()
}

func (self *Cron) Stop() error {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.cron.Stop()

	var result error

	for _, job := range self.jobs {
		if err := job.Stop(); err != nil {
			err = multierror.Append(result, err)
		}
	}

	return result
}

func (self *Cron) Clear() error {
	if err := self.Stop(); err != nil {
		return err
	}

	self.jobs = nil
	self.cron = cron.New()

	return nil
}

func (self *Cron) Add(config CronJobConfig) error {

	if config.Command == "" && config.Script == "" {
		return errors.New("invalid job config")
	}

	self.lock.Lock()
	defer self.lock.Unlock()
	job := &wrapper{NewJob(config, uuid.New()), self.c, self.tmp}

	if err := self.cron.AddJob(config.Cron, job); err != nil {
		return err
	}

	self.c <- &LoadedEvent{job.job.id, job.job.config.Name}

	self.jobs = append(self.jobs, job)

	return nil
}

func (self *Cron) AddFile(path string) error {
	return self.loadFromFile(path)
}

func (self *Cron) loadFromFile(path string) error {

	ext := filepath.Ext(path)

	var configs []CronJobConfig
	var err error
	switch ext {
	case ".yml", ".yaml":
		err = configFromYaml(path, &configs)
	case ".json":
		err = configFromJSON(path, &configs)
	default:
		return errors.New("not a valid cron file")
	}

	if err != nil {
		return err
	}

	for _, config := range configs {
		if err := self.Add(config); err != nil {
			return err
		}
	}

	return nil

}

func NewCron() *Cron {

	temp, err := ioutil.TempDir("", "gcron")
	if err != nil {
		return nil
	}

	return &Cron{
		cron: cron.New(),
		tmp:  temp,
		c:    make(chan TaskEvent, 10),
	}
}

func configFromYaml(path string, v *[]CronJobConfig) error {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(bs, v)
	if err != nil {
		var vv CronJobConfig
		if err = yaml.Unmarshal(bs, &vv); err != nil {
			return err
		}

		*v = []CronJobConfig{vv}
	}
	return nil
}

func configFromJSON(path string, v *[]CronJobConfig) error {

	bs, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bs, v)
	if err != nil {
		var vv CronJobConfig
		if err = json.Unmarshal(bs, &vv); err != nil {
			return err
		}
		*v = []CronJobConfig{vv}
	}
	return nil
}
