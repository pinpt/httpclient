#
# Makefile for building all things related to this repo
#
NAME := httpclient
ORG := pinpt
PKG := $(ORG)/$(NAME)
PROG_NAME := $(NAME)

SHELL := /bin/bash
BASEDIR := $(shell echo $${PWD})

.PHONY: default test

default: test

test:
	go test -v -cover ./