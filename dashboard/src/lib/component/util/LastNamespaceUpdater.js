import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import { withApollo } from "react-apollo";

const mutation = gql`
  mutation SetLastNamespace($name: String!) {
    setLastNamespace(name: $name) @client
  }
`;

class LastNamespaceUpdater extends React.PureComponent {
  static propTypes = {
    client: PropTypes.object.isRequired,
    namespace: PropTypes.string.isRequired,
  };

  componentDidMount() {
    this.updateLastNamespace();
  }

  componentDidUpdate() {
    this.updateLastNamespace();
  }

  updateLastNamespace = () => {
    const { namespace } = this.props;
    const variables = { name: namespace };

    this.props.client.mutate({ mutation, variables });
  };

  render() {
    return null;
  }
}

export default withApollo(LastNamespaceUpdater);
