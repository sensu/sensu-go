import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import { withApollo } from "react-apollo";

const mutation = gql`
  mutation SetLastEnvironment($org: String!, $env: String!) {
    setLastEnvironment(organization: $org, environment: $env) @client
  }
`;

class LastEnvironmentUpdater extends React.PureComponent {
  static propTypes = {
    client: PropTypes.object.isRequired,
    environment: PropTypes.string.isRequired,
    organization: PropTypes.string.isRequired,
  };

  componentDidMount() {
    this.updateLastEnvironment();
  }

  componentDidUpdate() {
    this.updateLastEnvironment();
  }

  updateLastEnvironment = () => {
    const { organization: org, environment: env } = this.props;
    const variables = { org, env };

    this.props.client.mutate({ mutation, variables });
  };

  render() {
    return null;
  }
}

export default withApollo(LastEnvironmentUpdater);
