import React from "react";
import PropTypes from "prop-types";
import { withRouter, matchShape, Link } from "found";

export function makeNamespacedPath({ org, env }) {
  return path => (path[0] === "/" ? path : `/${org}/${env}/${path}`);
}

class NamespaceLink extends React.Component {
  static propTypes = {
    children: PropTypes.node.isRequired,
    match: matchShape.isRequired,
    to: PropTypes.string.isRequired,
  };

  render() {
    const { to, match, children, ...props } = this.props;
    const path = makeNamespacedPath(match.params)(to);

    return (
      <Link to={path} {...props}>
        {children}
      </Link>
    );
  }
}

export default withRouter(NamespaceLink);
