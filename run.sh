#! /bin/bash

export DATABASE_HOST=ted.ck5mrdxgowlk.ap-northeast-2.rds.amazonaws.com
export DATABASE_PASSWORD=ePix9L5ILrw3
export DATABASE_NAME=reservation


go test -v ./...
go run -v main.go