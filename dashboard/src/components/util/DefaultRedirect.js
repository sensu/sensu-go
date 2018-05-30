import React from "react";
import PropTypes from "prop-types";
import { Redirect } from "react-router-dom";
import { graphql } from "react-apollo";
import gql from "graphql-tag";

class DefaultRedirect extends React.PureComponent {
  static propTypes = {
    data: PropTypes.object.isRequired,
  };

  render() {
    const { data } = this.props;

    // TODO: Store and retrieve last viewed environment.
    const lastEnvironment = "/default/default";

    const nextPath = data.auth.accessToken ? lastEnvironment : "/signin";

    return <Redirect to={nextPath} />;
  }
}

export default graphql(gql`
  query {
    auth @client {
      accessToken
    }
  }
`)(DefaultRedirect);
