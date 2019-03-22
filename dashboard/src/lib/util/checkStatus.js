const success = "success";
const warning = "warning";
const critical = "critical";
const unknown = "unknown";

const codeMap = {
  0: success,
  1: warning,
  2: critical,
};

// Given exit status return identifier.
// 0 == ok
// 1 == warning
// 2 == error
// > 2 == unknown
// eslint-disable-next-line import/prefer-default-export
export function statusCodeToId(code) {
  const st = codeMap[code];
  if (st) {
    return st;
  }
  return unknown;
}
