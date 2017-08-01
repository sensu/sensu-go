.PHONY: default clean sensu_backend services hooks packages rpms debs

##
# Prerequesite checks
##

# ensure GOARCH & GOOS are set and are valid

##
# FPM
##
NAME=sensu-backend
VERSION=$(shell cat version/version.txt)
ITERATION=$(shell cat version/iteration.txt)
ARCHITECTURE=$(GOARCH)
DESCRIPTION="Sensu is a monitoring thing"
LICENSE=MIT
VENDOR="Sensu, Inc."
MAINTAINER="Sensu Support <support@sensu.io>"
URL="https://sensuapp.org"

BIN_NAME=sensu-backend
BIN_TARGET_PATH=/usr/bin/$(BIN_NAME)
BIN_TARGET_PATH_ALT=/usr/local/bin/$(BIN_NAME)
BIN_SOURCE_PATH=target/$(TARGET_OS)-$(TARGET_ARCH)/$(BIN_NAME)

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

##
# Services
##
SERVICE_NAME=sensu-$(COMPONENT)
SERVICE_USER=sensu
SERVICE_GROUP=sensu
SERVICE_USER_MAC=_$(SERVICE_USER)
SERVICE_GROUP_MAC=_$(SERVICE_GROUP)
SERVICE_COMMAND_PATH=$(BIN_TARGET_PATH)
SERVICE_COMMAND_PATH_ALT=$(BIN_TARGET_PATH_ALT)
SERVICE_COMMAND_PATH_ALT=$(BIN_TARGET_PATH_ALT)
SERVICE_COMMAND_ARGUMENTS="start"

##
# Hooks
##
HOOKS_PATH=packaging/hooks
HOOKS_VALUES=service=sensu-backend
HOOKS_VALUES+= prefix=$(HOOKS_PATH)/common
HOOKS_VALUES+= common_files=os-functions,group-functions,user-functions

##
# Targets
##
default: all

all: sensu_backend

clean:
	rm -r build

sensu_backend: services hooks packages

services: service_sysvinit service_systemd service_launchd

service_sysvinit:
	pleaserun -p sysv --overwrite --no-install-actions \
	--install-prefix packaging/services/sysv \
	--user $(SERVICE_USER) --group $(SERVICE_GROUP) \
	$(SERVICE_COMMAND_PATH) $(SERVICE_COMMAND_ARGS)

service_systemd:
	pleaserun -p systemd --overwrite --no-install-actions \
	--install-prefix packaging/services/systemd \
	--user $(SERVICE_USER) --group $(SERVICE_GROUP) \
	$(SERVICE_COMMAND_PATH) $(SERVICE_COMMAND_ARGS)

service_launchd:
	pleaserun -p launchd --overwrite --no-install-actions \
	--install-prefix packaging/services/launchd \
	--user $(SERVICE_USER_MAC) --group $(SERVICE_GROUP_MAC) \
	$(SERVICE_COMMAND_PATH_ALT) $(SERVICE_COMMAND_ARGS)

hooks: hooks_deb

hooks_deb:
	erb $(HOOKS_VALUES) $(HOOKS_PATH)/deb/preinst.erb > $(HOOKS_PATH)/deb/preinst
	erb $(HOOKS_VALUES) $(HOOKS_PATH)/deb/postinst.erb > $(HOOKS_PATH)/deb/postinst
	erb $(HOOKS_VALUES) $(HOOKS_PATH)/deb/prerm.erb > $(HOOKS_PATH)/deb/prerm
	erb $(HOOKS_VALUES) $(HOOKS_PATH)/deb/postrm.erb > $(HOOKS_PATH)/deb/postrm

hooks_rpm:
	erb $(HOOKS_VALUES) $(HOOKS_PATH)/rpm/pre.erb > $(HOOKS_PATH)/rpm/pre
	erb $(HOOKS_VALUES) $(HOOKS_PATH)/rpm/post.erb > $(HOOKS_PATH)/rpm/post
	erb $(HOOKS_VALUES) $(HOOKS_PATH)/rpm/preun.erb > $(HOOKS_PATH)/rpm/preun
	erb $(HOOKS_VALUES) $(HOOKS_PATH)/rpm/postun.erb > $(HOOKS_PATH)/rpm/postun

packages: debs rpms

debs:
	deb_sysvinit
	deb_systemd

deb_sysvinit:
	#fpm $(FPM_FLAGS) $(FILES_MAP)

deb_systemd:
	#fpm $(FPM_FLAGS) $(FILES_MAP)

rpms:
	rpm_sysvinit
	rpm_systemd

rpm_sysvinit:
	#fpm $(FPM_FLAGS) \
	# --architecture $(shell packaging/rpm/safe-architecture.sh $(ARCHITECTURE)) \
	# $(FILES_MAP)

rpm_systemd:
	#fpm $(FPM_FLAGS) \
	# --architecture $(shell packaging/rpm/safe-architecture.sh $(ARCHITECTURE)) \
	# $(FILES_MAP)
