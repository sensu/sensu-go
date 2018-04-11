import React from "react";
import PropTypes from "prop-types";
import { compose, mapProps, hoistStatics } from "recompose";
import { withRouter, Link } from "found";

export function makeNamespacedPath({ organization, environment }) {
  return path =>
    path[0] === "/" ? path : `/${organization}/${environment}/${path}`;
}

export const namespaceShape = PropTypes.shape({
  organization: PropTypes.string.isRequired,
  environment: PropTypes.string.isRequired,
});

export const withNamespace = hoistStatics(
  compose(
    withRouter,
    mapProps(({ match, ...props }) => ({
      currentNamespace: {
        organization: match.params.organization,
        environment: match.params.environment,
      },
      ...props,
    })),
  ),
);

class NamespaceLink extends React.Component {
  static propTypes = {
    children: PropTypes.node.isRequired,
    currentNamespace: namespaceShape.isRequired,
    to: PropTypes.string.isRequired,
  };

  render() {
    const { to, currentNamespace, children, ...props } = this.props;
    const path = makeNamespacedPath(currentNamespace)(to);

    return (
      <Link to={path} {...props}>
        {children}
      </Link>
    );
  }
}

export default withNamespace(NamespaceLink);
