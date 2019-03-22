import React from "react";
import { Route, Redirect } from "react-router-dom";
import { redirectKey } from "/lib/constant/queryParams";

const signinPath = "/signin";

class SigninRedirect extends React.PureComponent {
  renderRedirect = ({ location }) => {
    let queryParams = location.search || "?";

    // Add next path
    if (location.pathname !== signinPath) {
      const newQuery = new URLSearchParams(queryParams);
      newQuery.set(redirectKey, location.pathname + location.search);
      queryParams = `?${newQuery.toString()}`;
    }

    return <Redirect to={`${signinPath}${queryParams}`} />;
  };

  render() {
    return <Route render={this.renderRedirect} />;
  }
}

export default SigninRedirect;
