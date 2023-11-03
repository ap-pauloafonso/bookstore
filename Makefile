include .env

docker-run:
	docker-compose up --build

docker-stop:
	docker-compose down

local-run:
	go run .

gen-docs:
	swag init

test-unit:
	go test ./... -count=1 --cover  -short -coverprofile=coverage-unit.out

test-all:
	go test ./... -count=1 --cover -skip Test_E2E -coverprofile=coverage.out

test-e2e:
	 go test -v -run Test_E2E

coverage-cleanup:
	rm -f coverage.out
	rm -f coverage-unit.out

coverage-unit: coverage-cleanup test-unit
	go tool cover -html=coverage-unit.out

coverage-all: coverage-cleanup test-all
	go tool cover -html=coverage.out