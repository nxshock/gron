package main

import (
	"fmt"
)

type RestartRule int

const (
	No RestartRule = iota
	OnError
)

func (r *RestartRule) MarshalText() (text []byte, err error) {
	switch *r {
	case No:
		return []byte("no"), nil
	case OnError:
		return []byte("on-error"), nil
	}

	return nil, fmt.Errorf("unknown restart rule: %v", r)
}

func (r *RestartRule) UnmarshalText(text []byte) error {
	switch string(text) {
	case "no":
		*r = No
		return nil
	case "on-error":
		*r = OnError
		return nil
	}

	return fmt.Errorf("unknown restart rule: %s", string(text))
}
