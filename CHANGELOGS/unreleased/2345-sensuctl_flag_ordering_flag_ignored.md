### Fixed
- Fixed a bug in `sensuctl` where global/persistent flags, such as `--namespace`
  and `--config-dir`, would get ignored if they were passed after a sub-command
  local flag, such as `--format`.
