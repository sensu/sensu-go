let _nextId = 0;

export default () => {
  const current = _nextId;
  _nextId += 1;
  return `${current}`;
};
