#!/bin/bash

go test -bench="."
go test -bench="." -tags new

