TARGET_OS=linux
TARGET_ARCH=amd64

# Load version & iteration from file
VERSION=1.0.0
ITERATION=1

NAME=sensu-$(COMPONENT)
ARCHITECTURE=$(GOARCH)
DESCRIPTION="Sensu is a monitoring thing"
LICENSE=MIT
VENDOR="Sensu, Inc."
MAINTAINER="Sensu Support <support@sensu.io>"
URL="https://sensuapp.org"

SERVICE_NAME=sensu-$(COMPONENT)
SERVICE_USER=sensu
SERVICE_GROUP=sensu

BINARY_NAME=sensu-backend
BINARY_START_ARGUMENTS="start"

# /usr/bin : linux
# /usr/local/bin : macOS, freebsd, openbsd, solaris, maybe aix?
# C:\Program Files\Sensu\sensu-agent\bin : windows (drive should be chooseable by user)
BINARY_TARGET_PATH=/usr/bin/$(BINARY_NAME)
BINARY_SOURCE_PATH=target/$(TARGET_OS)-$(TARGET_ARCH)/$(BINARY_NAME)

FILES_MAP = \
	$(BINARY_SOURCE_PATH)=$(BINARY_TARGET_PATH)

FPM_FLAGS = \
	--input-type dir \
	--output-type deb \
	--name $(NAME) \
	--version $(VERSION) \
	--iteration $(ITERATION) \
	--architecture $(ARCHITECTURE) \
	--description $(DESCRIPTION) \
	--url $(URL) \
	--license $(LICENSE) \
	--vendor $(VENDOR) \
	--maintainer $(MAINTAINER)

PACKAGE_TYPES = \
	deb \
	rpm

ifeq ($(filter $(PACKAGE_TYPE),$(PACKAGE_TYPES)),)
    $(error PACKAGE_TYPE environment variable must be set to one of [$(PACKAGE_TYPES)])
endif

include packaging/$(PACKAGE_TYPE).mk

.PHONY: default
default: all

all: services hooks package

.PHONY: clean
clean:
	rm -r build

.PHONY: services
services:
	@echo "Not implemented yet"

.PHONY: hooks
hooks:
	@echo "Not implemented yet"

.PHONY: package
package:
	#fpm $(FPM_FLAGS) $(FILES_MAP)
