import React from "react";
import { routerShape } from "found";
import { getAccessToken } from "../utils/authentication";

class RootRedirectPage extends React.Component {
  static propTypes = {
    router: routerShape.isRequired,
  };

  componentWillMount() {
    const handleToken = token => {
      // TODO: Retrieve last environment
      let nextPath = "/default/default";
      if (token === null) {
        nextPath = "/login";
      }
      this.props.router.push(nextPath);
    };

    const handleTokenError = () => {
      this.props.router.push("/login");
    };

    getAccessToken().then(handleToken, handleTokenError);
  }

  render() {
    return null;
  }
}

export default RootRedirectPage;
