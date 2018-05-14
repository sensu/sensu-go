export default () => {
  const values = ["production", "development", "test"];

  if (!values.includes(process.env.NODE_ENV)) {
    throw new Error(
      `The NODE_ENV environment variable must be one of "${[
        values.slice(0, -1).join('", "'),
      ]
        .concat(values.slice(-1))
        .join('", or "')}".`,
    );
  }
};
