ifneq ("$(wildcard '.env')","")
include .env
endif

lint:
	sh scripts/lint.sh
