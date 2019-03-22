const success = "success";
const warning = "warning";
const critical = "critical";
const unknown = "unknown";

// Given exit status return identifier.
// 0 == ok
// 1 == warning
// 2 == error
// > 2 == unknown
// eslint-disable-next-line import/prefer-default-export
export function statusCodeToId(st) {
  if (st === 0) {
    return success;
  } else if (st === 1) {
    return warning;
  } else if (st === 2) {
    return critical;
  }
  return unknown;
}
