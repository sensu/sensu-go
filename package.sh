#!/bin/sh

# TODO: figure out if FPM still sets a default epoch

TARGET_OS=linux
TARGET_ARCH=amd64

PACKAGE_NAME=sensu-agent
PACKAGE_VERSION=1.0.0.alpha.1
PACKAGE_ITERATION=1
PACKAGE_ARCH=$TARGET_ARCH # TODO: need helpers to do translation, e.g. amd64 deb, x86_64 rpm
PACKAGE_DESCRIPTION="Sensu is a monitoring thing"
PACKAGE_LICENSE=MIT
PACKAGE_VENDOR="Sensu, Inc."
PACKAGE_DEB_CATEGORY=net
PACKAGE_MAINTAINER="Sensu Support <support@sensu.io>"
PACKAGE_URL="https://sensuapp.org"

SERVICE_NAME=sensu-agent
SERVICE_USER=sensu
SERVICE_GROUP=sensu

BINARY_NAME=sensu-agent
BINARY_START_ARGS="start"

# /usr/bin : linux
# /usr/local/bin : macOS, freebsd, openbsd, solaris, maybe aix?
# C:\Program Files\Sensu\sensu-agent\bin : windows (drive should be chooseable by user)
BINARY_TARGET_PATH=/usr/bin/$BINARY_NAME
BINARY_SOURCE_PATH=target/$TARGET_OS-$TARGET_ARCH/$BINARY_NAME

# safe_rpm_version will return a version string that is rpm-compatible
# e.g.
#
safe_rpm_version () {
    echo "Not implemented yet"
    exit 1
}

# safe_debian_version will return a version string that is debian-compatible
# e.g. 1.0.0alpha1, 1.0.0beta3
# https://www.debian.org/doc/debian-policy/ch-controlfields.html#s-f-Version
safe_debian_version () {
    echo "Not implemented yet"
    exit 1
}

# safe_freebsd_version will return a version string that is freebsd-compatible
# e.g. 1.0.0a1, 1.0.0b3
# https://www.freebsd.org/doc/en/books/porters-handbook/makefile-naming.html#porting-pkgname-format
safe_freebsd_version () {
    echo "Not implemented yet"
    exit 1
}

generate_service_files () {
    for platform in sysv systemd launchd; do
	if [ platform = launchd ]; then
	    service_user=_$SERVICE_USER
	else
	    service_user=$SERVICE_USER
	fi
	if [ platform = launchd ]; then
	    service_group=_$SERVICE_GROUP
	else
	    service_group=$SERVICE_GROUP
	fi
	pleaserun -p $platform --overwrite --no-install-actions \
		  --install-prefix packaging/services/$platform \
		  --user $service_user --group $service_group \
		  $BINARY_TARGET $BINARY_START_ARGS
    done
}

mkdir -p packages/sysvinit
mkdir -p packages/systemd

# deb - sysvinit
fpm --input-type dir \
    --output-type deb \
    --name $PACKAGE_NAME \
    --version $PACKAGE_VERSION \
    --iteration $PACKAGE_ITERATION \
    --architecture $PACKAGE_ARCH \
    --package "packages/sysvinit/${PACKAGE_NAME}_${PACKAGE_VERSION}-${PACKAGE_ITERATION}_${PACKAGE_ARCH}.deb" \
    --description "$PACKAGE_DESCRIPTION" \
    --url "$PACKAGE_URL" \
    --license "$PACKAGE_LICENSE" \
    --vendor "$PACKAGE_VENDOR" \
    --category "$PACKAGE_DEB_CATEGORY" \
    --maintainer "$PACKAGE_MAINTAINER" \
    --deb-priority extra \
    --deb-init packaging/services/sysv/etc/init.d/$SERVICE_NAME \
    --deb-default packaging/services/sysv/etc/default/$SERVICE_NAME \
    $BINARY_SOURCE_PATH=$BINARY_TARGET_PATH

# deb - systemd
fpm --input-type dir \
    --output-type deb \
    --name $PACKAGE_NAME \
    --version $PACKAGE_VERSION \
    --iteration $PACKAGE_ITERATION \
    --architecture $PACKAGE_ARCH \
    --package "packages/systemd/${PACKAGE_NAME}_${PACKAGE_VERSION}-${PACKAGE_ITERATION}_${PACKAGE_ARCH}.deb" \
    --description "$PACKAGE_DESCRIPTION" \
    --url "$PACKAGE_URL" \
    --license "$PACKAGE_LICENSE" \
    --vendor "$PACKAGE_VENDOR" \
    --category "$PACKAGE_DEB_CATEGORY" \
    --maintainer "$PACKAGE_MAINTAINER" \
    --deb-priority extra \
    --deb-systemd packaging/services/systemd/etc/systemd/system/$SERVICE_NAME.service \
    --deb-default packaging/services/systemd/etc/default/$SERVICE_NAME \
    $BINARY_SOURCE_PATH=$BINARY_TARGET_PATH

#

# rpm - sysvinit
fpm --input-type dir \
    --output-type rpm \
    --name $PACKAGE_NAME \
    --version $PACKAGE_VERSION \
    --iteration $PACKAGE_ITERATION \
    --architecture $PACKAGE_ARCH \
    --package "packages/sysvinit/${PACKAGE_NAME}-${PACKAGE_VERSION}-${PACKAGE_ITERATION}.${PACKAGE_ARCH}.rpm" \
    --description "$PACKAGE_DESCRIPTION" \
    --url "$PACKAGE_URL" \
    --license "$PACKAGE_LICENSE" \
    --vendor "$PACKAGE_VENDOR" \
    --maintainer "$PACKAGE_MAINTAINER" \
    --rpm-init packaging/services/sysv/etc/init.d/$SERVICE_NAME \
    $BINARY_SOURCE_PATH=$BINARY_TARGET_PATH \
    packaging/services/sysv/etc/default/$SERVICE_NAME=/etc/default/

# rpm - systemd
fpm --input-type dir \
    --output-type rpm \
    --name $PACKAGE_NAME \
    --version $PACKAGE_VERSION \
    --iteration $PACKAGE_ITERATION \
    --architecture $PACKAGE_ARCH \
    --package "packages/systemd/${PACKAGE_NAME}-${PACKAGE_VERSION}-${PACKAGE_ITERATION}.${PACKAGE_ARCH}.rpm" \
    --description "$PACKAGE_DESCRIPTION" \
    --url "$PACKAGE_URL" \
    --license "$PACKAGE_LICENSE" \
    --vendor "$PACKAGE_VENDOR" \
    --maintainer "$PACKAGE_MAINTAINER" \
    $BINARY_SOURCE_PATH=$BINARY_TARGET_PATH \
    packaging/services/systemd/etc/systemd/system/$SERVICE_NAME.service=/lib/systemd/system/ \
    packaging/services/systemd/etc/default/$SERVICE_NAME=/etc/default/
