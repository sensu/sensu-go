import React from "react";
import PropTypes from "prop-types";
import { Link, Route } from "react-router-dom";

class NamespaceLink extends React.PureComponent {
  static propTypes = {
    ...Link.PropTypes,
    to: PropTypes.string.isRequired,
    namespace: PropTypes.shape({
      organization: PropTypes.string.isRequired,
      environment: PropTypes.string.isRequired,
    }),
  };

  static defaultProps = {
    namespace: undefined,
  };

  renderLink(organization, environment) {
    const { to, namespace: _namespace, ...props } = this.props;
    return <Link {...props} to={`/${organization}/${environment}${to}`} />;
  }

  render() {
    const { namespace } = this.props;

    if (this.props.namespace) {
      return this.renderLink(namespace.organization, namespace.environment);
    }

    return (
      <Route
        path="/:organization/:environment"
        render={({
          match: {
            params: { organization, environment },
          },
        }) => this.renderLink(organization, environment)}
      />
    );
  }
}

export default NamespaceLink;
