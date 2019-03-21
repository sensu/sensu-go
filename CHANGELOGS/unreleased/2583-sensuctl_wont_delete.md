### Fixed
- Fixed a bug in `sensuctl` where handlers and filters would only be deleted
from the default namespace, unless a `--namespace` flag was specified.
