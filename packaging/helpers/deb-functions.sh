#!/usr/bin/env bash

# safe_debian_version will return a version string that is debian-compatible
# e.g. 1.0.0alpha1, 1.0.0beta3
# https://www.debian.org/doc/debian-policy/ch-controlfields.html#s-f-Version
safe_debian_version() {
    echo "Not implemented yet"
    exit 1
}

DEB_FPM_FLAGS=(
    --input-type dir
    --output-type deb
    --name $PACKAGE_NAME
    --version $PACKAGE_VERSION
    --iteration $PACKAGE_ITERATION
    --architecture $PACKAGE_ARCH
    --package "packages/sysvinit/${PACKAGE_NAME}_${PACKAGE_VERSION}-${PACKAGE_ITERATION}_${PACKAGE_ARCH}.deb"
    --description "$PACKAGE_DESCRIPTION"
    --url "$PACKAGE_URL"
    --license "$PACKAGE_LICENSE"
    --vendor "$PACKAGE_VENDOR"
    --category "$PACKAGE_DEB_CATEGORY"
    --maintainer "$PACKAGE_MAINTAINER"
    --deb-priority extra
    --deb-init packaging/services/sysv/etc/init.d/$SERVICE_NAME
    --deb-default packaging/services/sysv/etc/default/$SERVICE_NAME
    --before-install packaging/hooks/deb/preinst
    --before-remove packaging/hooks/deb/prerm
    --after-install packaging/hooks/deb/postinst
    --after-remove packaging/hooks/deb/postrm
    $BINARY_SOURCE_PATH=$BINARY_TARGET_PATH
)
