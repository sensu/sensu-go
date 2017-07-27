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

. packaging/helpers/rpm-functions.sh
. packaging/helpers/deb-functions.sh
. packaging/helpers/freebsd-functions.sh

generate_services() {
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
		  $BINARY_TARGET_PATH $BINARY_START_ARGS
    done
}

generate_hooks() {
    hooks_path="packaging/hooks"
    common_path="${hooks_path}/common"
    prefix="packaging/hooks/common"
    common_files="os-functions,group-functions,user-functions"

    # render deb hooks
    erb service=$SERVICE_NAME prefix=$common_path common_files=$common_files $hooks_path/deb/preinst.erb > $hooks_path/deb/preinst
    erb service=$SERVICE_NAME prefix=$common_path common_files=$common_files $hooks_path/deb/postinst.erb > $hooks_path/deb/postinst
    erb service=$SERVICE_NAME prefix=$common_path common_files=$common_files $hooks_path/deb/prerm.erb > $hooks_path/deb/prerm
    erb service=$SERVICE_NAME prefix=$common_path common_files=$common_files $hooks_path/deb/postrm.erb > $hooks_path/deb/postrm

    # render rpm hooks
    erb service=$SERVICE_NAME prefix=$common_path common_files=$common_files $hooks_path/rpm/pre.erb > $hooks_path/rpm/pre
    erb service=$SERVICE_NAME prefix=$common_path common_files=$common_files $hooks_path/rpm/post.erb > $hooks_path/rpm/post
    erb service=$SERVICE_NAME prefix=$common_path common_files=$common_files $hooks_path/rpm/preun.erb > $hooks_path/rpm/preun
    erb service=$SERVICE_NAME prefix=$common_path common_files=$common_files $hooks_path/rpm/postun.erb > $hooks_path/rpm/postun
}

build_package() {
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
    --before-install packaging/hooks/deb/preinst \
    --before-remove packaging/hooks/deb/prerm \
    --after-install packaging/hooks/deb/postinst \
    --after-remove packaging/hooks/deb/postrm \
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
    --deb-default packaging/services/systemd/etc/default/$SERVICE_NAME \
    --before-install packaging/hooks/deb/preinst \
    --before-remove packaging/hooks/deb/prerm \
    --after-install packaging/hooks/deb/postinst \
    --after-remove packaging/hooks/deb/postrm \
    $BINARY_SOURCE_PATH=$BINARY_TARGET_PATH \
    packaging/services/systemd/etc/systemd/system/$SERVICE_NAME.service=/lib/systemd/system/

rpm_arch=$(safe_rpm_arch $PACKAGE_ARCH)

# rpm - sysvinit
fpm --input-type dir \
    --output-type rpm \
    --name $PACKAGE_NAME \
    --version $PACKAGE_VERSION \
    --iteration $PACKAGE_ITERATION \
    --architecture $rpm_arch \
    --package "packages/sysvinit/${PACKAGE_NAME}-${PACKAGE_VERSION}-${PACKAGE_ITERATION}.${rpm_arch}.rpm" \
    --description "$PACKAGE_DESCRIPTION" \
    --url "$PACKAGE_URL" \
    --license "$PACKAGE_LICENSE" \
    --vendor "$PACKAGE_VENDOR" \
    --maintainer "$PACKAGE_MAINTAINER" \
    --rpm-init packaging/services/sysv/etc/init.d/$SERVICE_NAME \
    --before-install packaging/hooks/rpm/pre \
    --before-remove packaging/hooks/rpm/preun \
    --after-install packaging/hooks/rpm/post \
    --after-remove packaging/hooks/rpm/postun \
    $BINARY_SOURCE_PATH=$BINARY_TARGET_PATH \
    packaging/services/sysv/etc/default/$SERVICE_NAME=/etc/default/

# rpm - systemd
fpm --input-type dir \
    --output-type rpm \
    --name $PACKAGE_NAME \
    --version $PACKAGE_VERSION \
    --iteration $PACKAGE_ITERATION \
    --architecture $rpm_arch \
    --package "packages/systemd/${PACKAGE_NAME}-${PACKAGE_VERSION}-${PACKAGE_ITERATION}.${rpm_arch}.rpm" \
    --description "$PACKAGE_DESCRIPTION" \
    --url "$PACKAGE_URL" \
    --license "$PACKAGE_LICENSE" \
    --vendor "$PACKAGE_VENDOR" \
    --maintainer "$PACKAGE_MAINTAINER" \
    --before-install packaging/hooks/rpm/pre \
    --before-remove packaging/hooks/rpm/preun \
    --after-install packaging/hooks/rpm/post \
    --after-remove packaging/hooks/rpm/postun \
    $BINARY_SOURCE_PATH=$BINARY_TARGET_PATH \
    packaging/services/systemd/etc/systemd/system/$SERVICE_NAME.service=/usr/lib/systemd/system/ \
    packaging/services/systemd/etc/default/$SERVICE_NAME=/etc/default/
}

case "$1" in
    services)
	generate_services
	;;

    hooks)
	generate_hooks
	;;

    build)
	generate_services
	generate_hooks
	build_package
	;;

    *)
	echo "$0: Invalid option."
	echo "Valid options: services, hooks, build"
	exit 1
	;;
esac
