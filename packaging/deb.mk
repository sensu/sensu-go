FPM_FLAGS += \
	--category net \
	--deb-priority extra \

SERVICE_TYPES = \
	sysvinit \
	systemd

ifneq ($(filter $(SERVICE_TYPE),$(SERVICE_TYPES)),)
  ifeq ($(SERVICE_TYPE), sysvinit)
    FPM_FLAGS += \
	--deb-init packaging/services/sysv/etc/init.d/$(SERVICE_NAME) \
	--deb-default packaging/services/sysv/etc/default/$(SERVICE_NAME)
  endif

  ifeq ($(SERVICE_TYPE), systemd)
    FPM_FLAGS += \
	--deb-default packaging/services/sysv/etc/default/$(SERVICE_NAME)

    FILES_MAP += \
	packaging/services/systemd/etc/systemd/system/$(SERVICE_NAME).service=/lib/systemd/system/
  endif
else
    $(error SERVICE_TYPE environment variable must be set to one of [$(SERVICE_TYPES)])
endif
