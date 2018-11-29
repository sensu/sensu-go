import React from "react";
import PropTypes from "prop-types";
import { withApollo } from "react-apollo";
import Button from "@material-ui/core/Button";
import Banner from "./Banner";
import retryLocalNetwork from "../../mutations/retryLocalNetwork";

class RetryConnectionBanner extends React.PureComponent {
  static propTypes = { client: PropTypes.object.isRequired };

  retryConnection = () => {
    console.log("this is running");
    retryLocalNetwork(this.props.client);
  };

  render() {
    return (
      <Banner
        message="You've lost network connection."
        variant="warning"
        actions={
          <Button color="inherit" onClick={() => this.retryConnection()}>
            reconnect
          </Button>
        }
      />
    );
  }
}

export default withApollo(RetryConnectionBanner);
