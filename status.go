package main

type Status int

const (
	Inactive Status = iota
	Running
	Error
	Restarting
)
