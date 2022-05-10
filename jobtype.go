package main

import (
	"fmt"
	"strings"
)

type JobType int

const (
	Cmd JobType = iota + 1
	Sql
)

func (j *JobType) MarshalText() (text []byte, err error) {
	switch int(*j) {
	case 1:
		return []byte("cmd"), nil
	case 2:
		return []byte("sql"), nil
	}

	return nil, fmt.Errorf("unknown job type: %v", j)
}

func (j *JobType) UnmarshalText(text []byte) error {
	switch strings.ToLower(string(text)) {
	case "cmd":
		*j = 1
		return nil
	case "sql":
		*j = 2
		return nil
	}

	return fmt.Errorf("unknown job type: %v", string(text))
}
