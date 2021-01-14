package main

import (
	"errors"
	consul "github.com/hashicorp/consul/api"
	"strings"
)

// Service Interface
type StringService interface {
	Uppercase(string) (string, error)
	Count(string) int
}

// Service Concrete
type stringService struct{
	consulAgent *consul.Agent
}

func (stringService) Uppercase(s string) (string, error) {
	if s == "" {
		return "", ErrEmpty
	}
	return strings.ToUpper(s), nil
}

func (stringService) Count(s string) int {
	return len(s)
}

var ErrEmpty = errors.New("empty string")