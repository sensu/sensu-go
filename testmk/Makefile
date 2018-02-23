.PHONY: default
default:
	@echo "default"

PRERELEASE=$(shell ../version/version.sh --prerelease)

ifneq ($(PRERELEASE),)
	@echo "prerelease"
else
	@echo "not a prerelease"
endif
