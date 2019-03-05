### Fixed

Fixed a bug where agent entity subscriptions would be communicated to the
backend incorrectly. Due to the scheduler using the subscriptions from the
HTTP header, this does not have any effect on scheduling.
