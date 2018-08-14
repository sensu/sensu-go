import React from "react";
import { withApollo } from "react-apollo";

import SigninDialog from "/components/partials/SigninDialog";

class SignInView extends React.Component {
  handleSuccess = () => {
    // ...
    console.info("handling success");
  };

  render() {
    return <SigninDialog hideBackdrop onSuccess={this.handleSuccess} />;
  }
}

export default withApollo(SignInView);
