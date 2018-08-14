import React from "react";
import { withApollo } from "react-apollo";

import SigninDialog from "/components/partials/SigninDialog";

class SignInView extends React.Component {
  render() {
    return <SigninDialog hideBackdrop />;
  }
}

export default withApollo(SignInView);
