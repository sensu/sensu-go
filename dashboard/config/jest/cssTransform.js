// This is a custom Jest transformer turning style imports into empty objects.
// http://facebook.github.io/jest/docs/tutorial-webpack.html

export function process() {
  return "module.exports = {};";
}
export function getCacheKey() {
  // The output is always the same.
  return "cssTransform";
}
