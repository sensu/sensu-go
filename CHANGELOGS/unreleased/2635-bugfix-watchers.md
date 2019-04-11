### Fixed
* Fixed an issue where etcd watchers were used incorrectly. This was causing
100% CPU usage in some components, as they would loop endlessly trying to get
results from watchers that broke, due to their stream terminating. Other
components would simply stop updating. Watchers now get reinstated when the
client regains connectivity.
