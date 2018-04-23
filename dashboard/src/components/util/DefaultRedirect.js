import React from "react";
import PropTypes from "prop-types";
import { Redirect } from "react-router-dom";

import withAuthTokens from "/components/util/withAuthTokens";

class DefaultRedirect extends React.PureComponent {
  static propTypes = {
    authTokens: PropTypes.object.isRequired,
  };

  render() {
    const { authTokens } = this.props;

    // TODO: Store and retrieve last viewed environment.
    const lastEnvironment = "/default/default";

    const nextPath = authTokens.accessToken ? lastEnvironment : "/login";

    return <Redirect to={nextPath} />;
  }
}

export default withAuthTokens(DefaultRedirect);
