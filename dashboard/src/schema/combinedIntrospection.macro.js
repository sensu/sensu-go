/* eslint-disable import/no-webpack-loader-syntax, import/extensions, import/no-unresolved */

/**
 * Emits results of introspection query using combined server and client
 * additions. In this way clients of the schema and any other tooling (GrapiQL
 * and friends) can be aware of client fields / types as well.
 */

import { buildASTSchema, introspectionFromSchema, parse } from "graphql";
import rawSchema from "!!raw-loader!./combined.graphql";

// Support legacy SDL spec; graphl-go support pending.
// https://github.com/graphql/graphql-js/blob/v0.13.0/src/language/parser.js#L89-L97
const parserOpts = { allowLegacySDLImplementsInterfaces: true };

export default ({ emitFile }) => {
  const schema = buildASTSchema(parse(rawSchema, parserOpts));
  const introspectionResults = { data: introspectionFromSchema(schema) };
  return emitFile("schema.json", JSON.stringify(introspectionResults));
};
