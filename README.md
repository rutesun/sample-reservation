# 회의실 예약 시스템

## Installation
패키지 관리 툴로 [godep](https://github.com/golang/dep) 을 사용
godep 설치 후 

`godep ensure`

## Test 
`go test -v ./...`

## Build
`go build -o ./app`

## Run
```
export DATABASE_HOST=ted.ck5mrdxgowlk.ap-northeast-2.rds.amazonaws.com
export DATABASE_PASSWORD=ePix9L5ILrw3
export DATABASE_NAME=reservation

./app
```
