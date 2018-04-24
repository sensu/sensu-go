import React from "react";
import PropTypes from "prop-types";

import ConditionalRoute from "/components/util/ConditionalRoute";
import withAuthTokens from "/components/util/withAuthTokens";

class UnauthenticatedRoute extends React.PureComponent {
  static propTypes = {
    ...ConditionalRoute.propTypes,
    authTokens: PropTypes.object.isRequired,
  };

  render() {
    const { authTokens, ...props } = this.props;

    return <ConditionalRoute {...props} active={!authTokens.accessToken} />;
  }
}

export default withAuthTokens(UnauthenticatedRoute);
