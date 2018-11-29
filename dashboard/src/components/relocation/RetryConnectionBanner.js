import React from "react";
import gql from "graphql-tag";
import InlineLink from "/components/InlineLink";
import Banner from "./Banner";

const mutation = gql`
  mutation RetryLocalNetworkMutation {
    retryLocalNetwork(retry: true) @client
  }
`;

class RetryConnectionBanner extends React.PureComponent {
  retryConnection = () => {
    console.log("It worked");
  };

  render() {
    return (
      <Banner
        // eslint-disable-next-line
        message="You've lost network connection."
        variant="warning"
        buttonMessage="Retry Connection"
        buttonAction={this.retryConnection}
      />
    );
  }
}

export default RetryConnectionBanner;
