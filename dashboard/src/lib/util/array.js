const shallowEqual = (as, bs) => {
  if (as === bs) {
    return true;
  }
  if (as.length !== bs.length) {
    return false;
  }
  for (let i = 0; i < as.length; i += 1) {
    if (as[i] !== bs[i]) {
      return false;
    }
  }
  return true;
};

export { shallowEqual };
