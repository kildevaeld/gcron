package internal

import "time"

type TaskEvent interface {
	JobID() string
	Name() string
}

type StartEvent struct {
	id   string
	name string
	time time.Time
}

func (self *StartEvent) JobID() string {
	return self.id
}

func (self *StartEvent) Name() string {
	return self.name
}

func (self *StartEvent) Time() time.Time {
	return self.time
}

type FinishEvent struct {
	id       string
	name     string
	duration time.Duration
	err      error
}

func (self *FinishEvent) Name() string {
	return self.name
}

func (self *FinishEvent) JobID() string {
	return self.id
}

func (self *FinishEvent) Duration() time.Duration {
	return self.duration
}

func (self *FinishEvent) Error() error {
	return self.err
}

type ErrorEvent struct {
	id       string
	name     string
	duration time.Duration
	err      error
}

func (self *ErrorEvent) Name() string {
	return self.name
}

func (self *ErrorEvent) JobID() string {
	return self.id
}

func (self *ErrorEvent) Error() error {
	return self.err
}

type LoadedEvent struct {
	id   string
	name string
}

func (self *LoadedEvent) Name() string {
	return self.name
}

func (self *LoadedEvent) JobID() string {
	return self.id
}
