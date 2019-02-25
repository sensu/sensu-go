function consecutivePairs([f, s, ...rest], acc = []) {
  const next = [...acc, [f, s]];
  if (rest.length < 2) {
    return next;
  }
  return consecutivePairs(rest, next);
}

export default consecutivePairs;
