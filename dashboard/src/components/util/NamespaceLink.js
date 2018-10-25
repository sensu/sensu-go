import React from "react";
import PropTypes from "prop-types";
import { Link, Route } from "react-router-dom";

class NamespaceLink extends React.PureComponent {
  static propTypes = {
    ...Link.PropTypes,
    to: PropTypes.string.isRequired,
    namespace: PropTypes.string.isRequired,
  };

  static defaultProps = {
    namespace: undefined,
  };

  renderLink(namespace) {
    const { to, namespace: _namespace, ...props } = this.props;
    return <Link {...props} to={`/${namespace}${to}`} />;
  }

  render() {
    const { namespace } = this.props;

    if (namespace) {
      return this.renderLink(namespace);
    }

    return (
      <Route
        path="/:namespace"
        render={({ match: { params } }) => this.renderLink(params.namespace)}
      />
    );
  }
}

export default NamespaceLink;
