import React from "react";
import PropTypes from "prop-types";
import { Link, Route } from "react-router-dom";

class NamespaceLink extends React.PureComponent {
  static propTypes = {
    ...Link.PropTypes,
    component: PropTypes.func,
    to: PropTypes.string.isRequired,
    namespace: PropTypes.shape({
      organization: PropTypes.string.isRequired,
      environment: PropTypes.string.isRequired,
    }),
  };

  static defaultProps = {
    namespace: undefined,
    component: Link,
  };

  renderLink = (organization, environment, props) => {
    const { component: Component, to, ...other } = props;
    return <Component {...other} to={`/${organization}/${environment}${to}`} />;
  };

  render() {
    const { namespace, ...props } = this.props;

    if (namespace) {
      return this.renderLink(
        namespace.organization,
        namespace.environment,
        props,
      );
    }

    return (
      <Route
        path="/:organization/:environment"
        render={({
          match: {
            params: { organization, environment },
          },
        }) => this.renderLink(organization, environment, props)}
      />
    );
  }
}

export default NamespaceLink;
