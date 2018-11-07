import React from "react";
import PropTypes from "prop-types";
import { Link, Route } from "react-router-dom";

class NamespaceLink extends React.PureComponent {
  static propTypes = {
    ...Link.PropTypes,
    component: PropTypes.func,
    to: PropTypes.string.isRequired,
    namespace: PropTypes.string.isRequired,
  };

  static defaultProps = {
    namespace: undefined,
    component: Link,
  };

  renderLink = (namespace, props) => {
    const { component: Component, to, ...other } = props;
    return <Component {...other} to={`/${namespace}${to}`} />;
  };

  render() {
    const { namespace, ...props } = this.props;

    if (namespace) {
      return this.renderLink(namespace, props);
    }

    return (
      <Route
        path="/:namespace"
        render={({ match: { params } }) =>
          this.renderLink(params.namespace, props)
        }
      />
    );
  }
}

export default NamespaceLink;
