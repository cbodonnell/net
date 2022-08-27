package net

type Client interface {
	Run() error
}

type Relay interface {
	Run() error
}

type Server interface {
	Run() error
}
