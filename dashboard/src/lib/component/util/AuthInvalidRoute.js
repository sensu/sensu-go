import React from "react";
import PropTypes from "prop-types";
import { graphql } from "react-apollo";
import gql from "graphql-tag";

import ConditionalRoute from "/lib/component/util/ConditionalRoute";

class AuthInvalidRoute extends React.PureComponent {
  static propTypes = {
    ...ConditionalRoute.propTypes,
    data: PropTypes.object.isRequired,
  };

  render() {
    const { data, ...props } = this.props;

    return <ConditionalRoute {...props} active={data.auth.invalid} />;
  }
}

export default graphql(gql`
  query AuthInvalidRouteQuery {
    auth @client {
      invalid
    }
  }
`)(AuthInvalidRoute);
