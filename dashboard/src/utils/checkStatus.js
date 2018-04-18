import { compose } from "ramda";

const ok = "ok";
const warning = "warning";
const error = "error";

const colorMap = {
  [ok]: "green",
  [warning]: "yellow",
  [error]: "red",
};

// Given exit status return identifier.
// 0 == ok
// 1 == warning
// > 1 == error
export function statusCodeToId(st) {
  if (st === 0) return ok;
  else if (st === 1) return warning;
  return error;
}

// Given identifier return associated color.
export function statusToColor(st) {
  return colorMap[st];
}

// Given exit status return associated color.
export const statusCodeToColor = compose(statusToColor, statusCodeToId);
