include packaging/rpm-$(GOARCH).mk

SERVICE_TYPES = \
	sysvinit \
	systemd

ifneq ($(filter $(SERVICE_TYPE),$(SERVICE_TYPES)),)
    ifeq ($(SERVICE_TYPE), sysvinit)
	FPM_FLAGS += \
		--rpm-init packaging/services/sysv/etc/init.d/$(SERVICE_NAME)

	FILES_MAP += \
		packaging/services/sysv/etc/default/$(SERVICE_NAME)=/etc/default/
    endif

    ifeq ($(SERVICE_TYPE), systemd)
	FILES_MAP += \
		packaging/services/systemd/etc/systemd/system/$(SERVICE_NAME).service=/lib/systemd/system/ \
		packaging/services/systemd/etc/default/$(SERVICE_NAME)=/etc/default/
    endif
else
    $(error SERVICE_TYPE environment variable must be set to one of [$(SERVICE_TYPES)])
endif
