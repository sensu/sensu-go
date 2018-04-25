import React from "react";
import PropTypes from "prop-types";
import { graphql } from "react-apollo";
import gql from "graphql-tag";

import ConditionalRoute from "/components/util/ConditionalRoute";

class UnauthenticatedRoute extends React.PureComponent {
  static propTypes = {
    ...ConditionalRoute.propTypes,
    data: PropTypes.object.isRequired,
  };

  render() {
    const { data, ...props } = this.props;

    return <ConditionalRoute {...props} active={!data.auth.accessToken} />;
  }
}

export default graphql(gql`
  query UnauthenticatedRouteQuery {
    auth @client {
      accessToken
    }
  }
`)(UnauthenticatedRoute);
