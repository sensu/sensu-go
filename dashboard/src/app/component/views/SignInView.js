import React from "react";
import { withApollo } from "react-apollo";

import SigninDialog from "/app/component/partial/SigninDialog";

class SignInView extends React.Component {
  render() {
    return <SigninDialog hideBackdrop />;
  }
}

export default withApollo(SignInView);
