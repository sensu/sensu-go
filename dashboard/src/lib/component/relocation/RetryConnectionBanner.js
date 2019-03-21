import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { withApollo, graphql } from "react-apollo";
import Button from "@material-ui/core/Button";

import retryLocalNetwork from "/lib/mutation/retryLocalNetwork";
import Banner from "./Banner";

class RetryConnectionBanner extends React.PureComponent {
  static propTypes = {
    client: PropTypes.object.isRequired,
    data: PropTypes.object.isRequired,
  };

  retryConnection = () => {
    retryLocalNetwork(this.props.client);
  };

  render() {
    const { data } = this.props;
    return (
      <Banner
        // TODO: make message better by linking to a troubleshooting guide
        // or use NetworkInformation API: https://developer.mozilla.org/en-US/docs/Web/API/NetworkInformation
        message="You are offline. Live updates are currently disabled."
        variant="warning"
        actions={
          <Button
            color="inherit"
            onClick={() => this.retryConnection()}
            disabled={data.localNetwork.retry}
          >
            reconnect
          </Button>
        }
      />
    );
  }
}

export default graphql(gql`
  query RetryConnectionBannerQuery {
    localNetwork @client {
      retry
    }
  }
`)(withApollo(RetryConnectionBanner));
