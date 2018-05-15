/* eslint-disable import/no-webpack-loader-syntax, import/extensions */

/**
 * Emits results of introspection query using combined server and client
 * additions. In this way clients of the schema and any other tooling (GrapiQL
 * and friends) can be aware of client fields / types as well.
 */

import { buildSchema, introspectionFromSchema } from "graphql";
import rawSchema from "!!raw-loader!./combined.graphql";

export default ({ emitFile }) => {
  const schema = buildSchema(rawSchema);
  const introspectionResults = { data: introspectionFromSchema(schema) };
  return emitFile("schema.json", JSON.stringify(introspectionResults));
};
