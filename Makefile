.PHONY: default clean
.PHONY: android darwin dragonfly freebsd linux netbsd openbsd plan9 solaris windows
.PHONY: sensu_agent sensu_backend sensu_cli
.PHONY: hooks packages rpms debs

ANDROID_ARCHITECTURES   := arm
DARWIN_ARCHITECTURES    := 386 amd64 # arm arm64
DRAGONFLY_ARCHITECTURES := amd64
FREEBSD_ARCHITECTURES   := 386 amd64 arm
LINUX_ARCHITECTURES     := 386 amd64 #arm arm64 mips mipsle mips64 mips64le ppc64 ppc64le
NETBSD_ARCHITECTURES    := 386 amd64 arm
OPENBSD_ARCHITECTURES   := 386 amd64 arm
PLAN9_ARCHITECTURES     := 386 amd64
SOLARIS_ARCHITECTURES   := amd64
WINDOWS_ARCHITECTURES   := 386 amd64

##
# Setup
##
$(shell mkdir -p out)
HOST_GOOS := $(shell go env GOOS)
HOST_GOARCH := $(shell go env GOARCH)
VERSION_CMD := go run ./version/cmd/version/version.go

##
# FPM
##
VERSION := $(shell $(VERSION_CMD) -v)
BUILD_TYPE := $(shell $(VERSION_CMD) -t)
PRERELEASE := $(shell $(VERSION_CMD) -p)
ITERATION := $(shell $(VERSION_CMD) -i)
ARCHITECTURE := $(GOARCH)
DESCRIPTION="A monitoring framework that aims to be simple, malleable, and scalable."
LICENSE=MIT
VENDOR="Sensu, Inc."
MAINTAINER="Sensu Support <support@sensu.io>"
URL="https://sensuapp.org"

BIN_SOURCE_DIR=target/$(GOOS)-$(GOARCH)

ifeq ($(BUILD_TYPE),stable)
FPM_VERSION=$(VERSION)
else ifeq ($(BUILD_TYPE),nightly)
VERSION=$(shell $(VERSION_CMD) -b)
DATE=$(shell date "+%Y%m%d")
FPM_VERSION=$(VERSION)~nightly+$(DATE)
else
FPM_VERSION=$(VERSION)~$(PRERELEASE)
endif

FPM_FLAGS = \
	--version $(FPM_VERSION) \
	--iteration $(ITERATION) \
	--url $(URL) \
	--license $(LICENSE) \
	--vendor $(VENDOR) \
	--maintainer $(MAINTAINER)

##
# Services
##
SERVICE_USER=sensu
SERVICE_GROUP=sensu
SERVICE_USER_MAC=_$(SERVICE_USER)
SERVICE_GROUP_MAC=_$(SERVICE_GROUP)

##
# Hooks
##
HOOKS_BASE_PATH=packaging/hooks
HOOKS_VALUES=prefix=$(HOOKS_BASE_PATH)/common
HOOKS_VALUES+= common_files=os-functions,group-functions,user-functions,other-functions

##
# Targets
##
default: all

all: linux

clean:
	rm -rf out/
	rm -rf target/
	git clean -dXf dashboard

##
# Operating system targets
##


#android: arm

darwin: export GOOS=darwin
darwin: export BIN_TARGET_DIR=/usr/local/bin
darwin: export PLATFORM_SERVICES=service_launchd
darwin:
	for arch in $(DARWIN_ARCHITECTURES); do \
	    make GOARCH=$$arch sensu; \
	done

#dragonfly: amd64

#freebsd: 386 amd64 arm

linux: export GOOS=linux
linux: export BIN_TARGET_DIR=/usr/bin
linux: export PLATFORM_SERVICES=service_sysvinit service_systemd
linux: export HOOKS := hooks_deb hooks_rpm
linux: export PACKAGERS := deb rpm
linux:
	for arch in $(LINUX_ARCHITECTURES); do \
	    make GOARCH=$$arch sensu; \
	done

#netbsd: 386 amd64 arm

#openbsd: 386 amd64 arm

#plan9: 386 amd64

#solaris: amd64

#windows: 386 amd64

##
# Sensu targets
##
sensu:
	make sensu_backend SERVICES="$(PLATFORM_SERVICES)"
	make sensu_agent SERVICES="$(PLATFORM_SERVICES)"
	make sensu_cli SERVICES="service_none"

sensu_agent: FPM_FLAGS+= --name sensu-agent
sensu_agent: FPM_FLAGS+= --description "Sensu agent description here"
sensu_agent: BIN_NAME=sensu-agent
sensu_agent: export SERVICE_NAME=sensu-agent
sensu_agent: SERVICE_COMMAND_PATH=$(BIN_TARGET_DIR)/sensu-agent
sensu_agent: SERVICE_COMMAND_ARGS=start
sensu_agent: FILES_MAP=$(BIN_SOURCE_DIR)/sensu-agent=$(BIN_TARGET_DIR)/sensu-agent
sensu_agent: FILES_MAP+= packaging/files/agent.yml.example=/etc/sensu/agent.yml.example
sensu_agent: build_agent hooks packages

sensu_backend: FPM_FLAGS+= --name sensu-backend
sensu_backend: FPM_FLAGS+= --description "Sensu backend description here"
sensu_backend: BIN_NAME=sensu-backend
sensu_backend: export SERVICE_NAME=sensu-backend
sensu_backend: SERVICE_COMMAND_PATH=$(BIN_TARGET_DIR)/sensu-backend
sensu_backend: SERVICE_COMMAND_ARGS=start -c /etc/sensu/backend.yml
sensu_backend: FILES_MAP=$(BIN_SOURCE_DIR)/sensu-backend=$(BIN_TARGET_DIR)/sensu-backend
sensu_backend: FILES_MAP+= packaging/files/backend.yml.example=/etc/sensu/backend.yml.example
sensu_backend: build_backend hooks packages

sensu_cli: FPM_FLAGS+= --name sensu-cli
sensu_cli: FPM_FLAGS+= --description "Sensu cli description here"
sensu_cli: BIN_NAME=sensuctl
sensu_cli: FILES_MAP=$(BIN_SOURCE_DIR)/sensuctl=$(BIN_TARGET_DIR)/sensuctl
sensu_cli: export SERVICE_NAME=sensu-cli
sensu_cli: build_cli hooks packages

##
# Compile targets
##
build_agent:
	GOOS=$(GOOS) GOARCH=$(GOARCH) ./build.sh build_agent

build_backend:
	GOOS=linux GOARCH=amd64 ./build.sh dashboard
	GOOS=$(GOOS) GOARCH=$(GOARCH) ./build.sh build_backend

build_cli:
	GOOS=$(GOOS) GOARCH=$(GOARCH) ./build.sh build_cli

##
# Hook targets
##
hooks: HOOKS_VALUES += service=$(SERVICE_NAME)
hooks: $(HOOKS)

hooks_deb: export HOOK_PACKAGER=deb
hooks_deb:
	make render_hooks

hooks_rpm: export HOOK_PACKAGER=rpm
hooks_rpm:
	make render_hooks

render_hooks: HOOKS_PATH=$(HOOKS_BASE_PATH)/$(SERVICE_NAME)/$(HOOK_PACKAGER)
render_hooks:
	erb $(HOOKS_VALUES) $(HOOKS_PATH)/before-install.erb > $(HOOKS_PATH)/before-install
	erb $(HOOKS_VALUES) $(HOOKS_PATH)/after-install.erb > $(HOOKS_PATH)/after-install
	erb $(HOOKS_VALUES) $(HOOKS_PATH)/before-remove.erb > $(HOOKS_PATH)/before-remove
	erb $(HOOKS_VALUES) $(HOOKS_PATH)/after-remove.erb > $(HOOKS_PATH)/after-remove

##
# Package targets
##
packages: $(PACKAGERS)

# deb
deb: FPM_INITIAL_FLAGS=-s dir -t deb
deb: FPM_FLAGS += --architecture $(shell packaging/deb/safe-architecture.sh $(GOARCH))
deb: FPM_FLAGS+= --before-install packaging/hooks/$(SERVICE_NAME)/deb/before-install
deb: FPM_FLAGS+= --after-install packaging/hooks/$(SERVICE_NAME)/deb/after-install
deb: FPM_FLAGS+= --before-remove packaging/hooks/$(SERVICE_NAME)/deb/before-remove
deb: FPM_FLAGS+= --after-remove packaging/hooks/$(SERVICE_NAME)/deb/after-remove
deb: $(addprefix deb_, $(SERVICES))

deb_service_none: FPM_FLAGS += --package out/deb/none/
deb_service_none:
	mkdir -p out/deb/none
	fpm $(FPM_INITIAL_FLAGS) $(FPM_FLAGS) $(FILES_MAP)

deb_service_sysvinit: FPM_FLAGS += --package out/deb/sysvinit/
deb_service_sysvinit: FILES_MAP += packaging/services/$(SERVICE_NAME)/sysvinit/etc/init.d/$(SERVICE_NAME)=/etc/init.d/$(SERVICE_NAME)
deb_service_sysvinit:
	mkdir -p out/deb/sysvinit
	fpm $(FPM_INITIAL_FLAGS) $(FPM_FLAGS) $(FILES_MAP)

deb_service_systemd: FPM_FLAGS += --package out/deb/systemd/
deb_service_systemd: FILES_MAP += packaging/services/$(SERVICE_NAME)/systemd/etc/systemd/system/$(SERVICE_NAME).service=/lib/systemd/system/$(SERVICE_NAME).service
deb_service_systemd:
	mkdir -p out/deb/systemd
	fpm $(FPM_INITIAL_FLAGS) $(FPM_FLAGS) $(FILES_MAP)

# rpm
rpm: FPM_INITIAL_FLAGS=-s dir -t rpm
rpm: FPM_FLAGS+= --architecture $(shell packaging/rpm/safe-architecture.sh $(GOARCH))
rpm: FPM_FLAGS+= --before-install packaging/hooks/$(SERVICE_NAME)/rpm/before-install
rpm: FPM_FLAGS+= --after-install packaging/hooks/$(SERVICE_NAME)/rpm/after-install
rpm: FPM_FLAGS+= --before-remove packaging/hooks/$(SERVICE_NAME)/rpm/before-remove
rpm: FPM_FLAGS+= --after-remove packaging/hooks/$(SERVICE_NAME)/rpm/after-remove
rpm: $(addprefix rpm_, $(SERVICES))

rpm_service_none: FPM_FLAGS += --package out/rpm/none/
rpm_service_none:
	mkdir -p out/rpm/none
	fpm $(FPM_INITIAL_FLAGS) $(FPM_FLAGS) $(FILES_MAP)

rpm_service_sysvinit: FPM_FLAGS += --package out/rpm/sysvinit/
rpm_service_sysvinit: FILES_MAP += packaging/services/$(SERVICE_NAME)/sysvinit/etc/init.d/$(SERVICE_NAME)=/etc/init.d/$(SERVICE_NAME)
rpm_service_sysvinit:
	mkdir -p out/rpm/sysvinit
	fpm $(FPM_INITIAL_FLAGS) $(FPM_FLAGS) $(FILES_MAP)

rpm_service_systemd: FPM_FLAGS += --package out/rpm/systemd/
rpm_service_systemd: FILES_MAP += packaging/services/$(SERVICE_NAME)/systemd/etc/systemd/system/$(SERVICE_NAME).service=/lib/systemd/system/$(SERVICE_NAME).service
rpm_service_systemd:
	mkdir -p out/rpm/systemd
	fpm $(FPM_INITIAL_FLAGS) $(FPM_FLAGS) $(FILES_MAP)

##
# publish targets
##

# every deb distro/version supported by packagecloud.io (2018-05-07)
DEB_SYSVINIT_DISTRO_VERSIONS=
#DEB_SYSVINIT_DISTRO_VERSIONS += ubuntu/warty
#DEB_SYSVINIT_DISTRO_VERSIONS += ubuntu/hoary
#DEB_SYSVINIT_DISTRO_VERSIONS += ubuntu/breezy
#DEB_SYSVINIT_DISTRO_VERSIONS += ubuntu/dapper
#DEB_SYSVINIT_DISTRO_VERSIONS += ubuntu/edgy
#DEB_SYSVINIT_DISTRO_VERSIONS += ubuntu/feisty
#DEB_SYSVINIT_DISTRO_VERSIONS += ubuntu/gutsy
#DEB_SYSVINIT_DISTRO_VERSIONS += ubuntu/hardy
#DEB_SYSVINIT_DISTRO_VERSIONS += ubuntu/intrepid
#DEB_SYSVINIT_DISTRO_VERSIONS += ubuntu/jaunty
#DEB_SYSVINIT_DISTRO_VERSIONS += ubuntu/karmic
#DEB_SYSVINIT_DISTRO_VERSIONS += ubuntu/lucid
#DEB_SYSVINIT_DISTRO_VERSIONS += ubuntu/maverick
#DEB_SYSVINIT_DISTRO_VERSIONS += ubuntu/natty
#DEB_SYSVINIT_DISTRO_VERSIONS += ubuntu/oneiric
#DEB_SYSVINIT_DISTRO_VERSIONS += ubuntu/precise
#DEB_SYSVINIT_DISTRO_VERSIONS += ubuntu/quantal
#DEB_SYSVINIT_DISTRO_VERSIONS += ubuntu/raring
#DEB_SYSVINIT_DISTRO_VERSIONS += ubuntu/saucy
DEB_SYSVINIT_DISTRO_VERSIONS += ubuntu/trusty
#DEB_SYSVINIT_DISTRO_VERSIONS += ubuntu/utopic
#DEB_SYSVINIT_DISTRO_VERSIONS += debian/etch
#DEB_SYSVINIT_DISTRO_VERSIONS += debian/lenny
#DEB_SYSVINIT_DISTRO_VERSIONS += debian/squeeze
#DEB_SYSVINIT_DISTRO_VERSIONS += debian/wheezy
#DEB_SYSVINIT_DISTRO_VERSIONS += raspbian/wheezy
#DEB_SYSVINIT_DISTRO_VERSIONS += elementaryos/jupiter
#DEB_SYSVINIT_DISTRO_VERSIONS += elementaryos/luna
#DEB_SYSVINIT_DISTRO_VERSIONS += elementaryos/freya
#DEB_SYSVINIT_DISTRO_VERSIONS += linuxmint/petria
#DEB_SYSVINIT_DISTRO_VERSIONS += linuxmint/qiana
#DEB_SYSVINIT_DISTRO_VERSIONS += linuxmint/rebecca
#DEB_SYSVINIT_DISTRO_VERSIONS += linuxmint/rafaela
#DEB_SYSVINIT_DISTRO_VERSIONS += linuxmint/rosa

DEB_SYSTEMD_DISTRO_VERSIONS=
#DEB_SYSTEMD_DISTRO_VERSIONS += ubuntu/vivid
#DEB_SYSTEMD_DISTRO_VERSIONS += ubuntu/wily
DEB_SYSTEMD_DISTRO_VERSIONS += ubuntu/xenial
#DEB_SYSTEMD_DISTRO_VERSIONS += ubuntu/yakkety
#DEB_SYSTEMD_DISTRO_VERSIONS += ubuntu/zesty
#DEB_SYSTEMD_DISTRO_VERSIONS += ubuntu/artful
DEB_SYSTEMD_DISTRO_VERSIONS += ubuntu/bionic
DEB_SYSTEMD_DISTRO_VERSIONS += debian/jessie
DEB_SYSTEMD_DISTRO_VERSIONS += debian/stretch
DEB_SYSTEMD_DISTRO_VERSIONS += debian/buster
#DEB_SYSTEMD_DISTRO_VERSIONS += raspbian/jessie
#DEB_SYSTEMD_DISTRO_VERSIONS += raspbian/stretch
#DEB_SYSTEMD_DISTRO_VERSIONS += raspbian/buster
#DEB_SYSTEMD_DISTRO_VERSIONS += linuxmint/sarah
#DEB_SYSTEMD_DISTRO_VERSIONS += linuxmint/serena
#DEB_SYSTEMD_DISTRO_VERSIONS += linuxmint/sonya
#DEB_SYSTEMD_DISTRO_VERSIONS += linuxmint/sylvia

DEB_DISTRO_VERSIONS:=$(DEB_SYSVINIT_DISTRO_VERSIONS) $(DEB_SYSTEMD_DISTRO_VERSIONS)

# every rpm distro/version supported by packagecloud.io (2018-05-07)
RPM_SYSVINIT_DISTRO_VERSIONS=
#RPM_SYSVINIT_DISTRO_VERSIONS += el/5
RPM_SYSVINIT_DISTRO_VERSIONS += el/6
#RPM_SYSVINIT_DISTRO_VERSIONS += ol/5
#RPM_SYSVINIT_DISTRO_VERSIONS += ol/6
#RPM_SYSVINIT_DISTRO_VERSIONS += scientific/5
#RPM_SYSVINIT_DISTRO_VERSIONS += scientific/6
#RPM_SYSVINIT_DISTRO_VERSIONS += sles/11.4
#RPM_SYSVINIT_DISTRO_VERSIONS += fedora/14
#RPM_SYSVINIT_DISTRO_VERSIONS += poky/jethro
#RPM_SYSVINIT_DISTRO_VERSIONS += poky/kogroth

RPM_SYSTEMD_DISTRO_VERSIONS=
RPM_SYSTEMD_DISTRO_VERSIONS += el/7
#RPM_SYSTEMD_DISTRO_VERSIONS += ol/7
#RPM_SYSTEMD_DISTRO_VERSIONS += scientific/7
#RPM_SYSTEMD_DISTRO_VERSIONS += sles/12.0
#RPM_SYSTEMD_DISTRO_VERSIONS += sles/12.1
#RPM_SYSTEMD_DISTRO_VERSIONS += sles/12.2
#RPM_SYSTEMD_DISTRO_VERSIONS += sles/12.3
#RPM_SYSTEMD_DISTRO_VERSIONS += opensuse/13.1
#RPM_SYSTEMD_DISTRO_VERSIONS += opensuse/13.2
#RPM_SYSTEMD_DISTRO_VERSIONS += opensuse/42.1
#RPM_SYSTEMD_DISTRO_VERSIONS += opensuse/42.2
#RPM_SYSTEMD_DISTRO_VERSIONS += opensuse/42.3
#RPM_SYSTEMD_DISTRO_VERSIONS += fedora/15
#RPM_SYSTEMD_DISTRO_VERSIONS += fedora/16
#RPM_SYSTEMD_DISTRO_VERSIONS += fedora/17
#RPM_SYSTEMD_DISTRO_VERSIONS += fedora/18
#RPM_SYSTEMD_DISTRO_VERSIONS += fedora/19
#RPM_SYSTEMD_DISTRO_VERSIONS += fedora/20
#RPM_SYSTEMD_DISTRO_VERSIONS += fedora/21
#RPM_SYSTEMD_DISTRO_VERSIONS += fedora/22
#RPM_SYSTEMD_DISTRO_VERSIONS += fedora/23
#RPM_SYSTEMD_DISTRO_VERSIONS += fedora/24
#RPM_SYSTEMD_DISTRO_VERSIONS += fedora/25
#RPM_SYSTEMD_DISTRO_VERSIONS += fedora/26
#RPM_SYSTEMD_DISTRO_VERSIONS += fedora/27

RPM_DISTRO_VERSIONS:=$(RPM_SYSVINIT_DISTRO_VERSIONS) $(RPM_SYSTEMD_DISTRO_VERSIONS)

SYSVINIT_DISTRO_VERSIONS:=$(DEB_SYSVINIT_DISTRO_VERSIONS) $(RPM_SYSVINIT_DISTRO_VERSIONS)
SYSTEMD_DISTRO_VERSIONS:=$(DEB_SYSTEMD_DISTRO_VERSIONS) $(RPM_SYSTEMD_DISTRO_VERSIONS)
ALL_DISTRO_VERSIONS:=$(DEB_DISTRO_VERSIONS) $(RPM_DISTRO_VERSIONS)

publish:
	make publish-noservice-packages
	make publish-sysvinit-packages
	make publish-systemd-packages

PC_PUSH_CMD=package_cloud push --skip-errors --config .packagecloud

publish_travis:
	make publish-noservice-packages
	make publish-sysvinit-packages
	make publish-systemd-packages

PC_PUSH_CMD=package_cloud push --skip-errors
PC_USER=sensu
PC_REPOSITORY=$(shell $(VERSION_CMD) -t)

##
# publish packages without a service
##
publish-noservice-packages: $(addprefix publish-noservice-package-,$(ALL_DISTRO_VERSIONS))

$(addprefix publish-noservice-package-,$(DEB_DISTRO_VERSIONS)):
	$(PC_PUSH_CMD) $(PC_USER)/$(PC_REPOSITORY)/$(subst publish-noservice-package-,,$@) out/deb/none/*

$(addprefix publish-noservice-package-,$(RPM_DISTRO_VERSIONS)):
	$(PC_PUSH_CMD) $(PC_USER)/$(PC_REPOSITORY)/$(subst publish-noservice-package-,,$@) out/rpm/none/*

##
# publish packages with sysvinit
##
publish-sysvinit-packages: $(addprefix publish-sysvinit-package-,$(SYSVINIT_DISTRO_VERSIONS))

$(addprefix publish-sysvinit-package-,$(DEB_SYSVINIT_DISTRO_VERSIONS)):
	$(PC_PUSH_CMD) $(PC_USER)/$(PC_REPOSITORY)/$(subst publish-sysvinit-package-,,$@) out/deb/sysvinit/*

$(addprefix publish-sysvinit-package-,$(RPM_SYSVINIT_DISTRO_VERSIONS)):
	$(PC_PUSH_CMD) $(PC_USER)/$(PC_REPOSITORY)/$(subst publish-sysvinit-package-,,$@) out/rpm/sysvinit/*

##
# publish packages with systemd
##
publish-systemd-packages: $(addprefix publish-systemd-package-,$(SYSTEMD_DISTRO_VERSIONS))

$(addprefix publish-systemd-package-,$(DEB_SYSTEMD_DISTRO_VERSIONS)):
	$(PC_PUSH_CMD) $(PC_USER)/$(PC_REPOSITORY)/$(subst publish-systemd-package-,,$@) out/deb/systemd/*

$(addprefix publish-systemd-package-,$(RPM_SYSTEMD_DISTRO_VERSIONS)):
	$(PC_PUSH_CMD) $(PC_USER)/$(PC_REPOSITORY)/$(subst publish-systemd-package-,,$@) out/rpm/systemd/*
