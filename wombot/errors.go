package main

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrNoAttack = Error("there aren't any attacks")
	ErrNoImgs   = Error("no required key in Images")
)
