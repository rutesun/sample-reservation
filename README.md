# 회의실 예약 시스템

## Installation
패키지 관리 툴로 [godep](https://github.com/golang/dep) 을 사용
godep 설치 후 

`godep ensure`
명령어를 통해 vendor 설치

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

## 문제해결 전략
### Packages
- config
    - 환경변수를 통해 어플리케이션 구동에 필수적인 값들을 주입받음
    
- controller
    - request 의 validation 을 체크
    - business logic 을 실행
    
- reservation
    - business logic + domain 
    - request 의 정합성, 유효성을 체크
    - 필수적인 use case를 interface로 정의하여 생성자에서 인자로 주입 받음
    - persistence layer와 logic layer의 결합도를 낮춤
      
- mariadb
    - business logic 에서 정의된 interface 를 구현
    - transaction 관리
    
- log
    - 기본적으로 stdout 으로 동작하며 io.Writer 를 주입 받는 형식으로 확장 가능
    - 기존 log 패키지 인터페이스를 확장하고 여러 3rd party library 와 쉽게 호환가능
    
- main
    - config 설정 정보를 만들고 config 설정 정보를 통해 db 객체를 생성 후 business layer에 주입
    - endpoint 와 controller 연결
    
    
### DB
회의실과 예약 정보는 정규화된 데이터인데 redis, memcached 같은 inmemory db 는 document 를 표현하고 
query에 제약 사항이 있어 mariadb를 사용함