import React from "react";
import { routerShape } from "found";
import { getAccessToken } from "../utils/authentication";

class RootRedirectPage extends React.Component {
  static propTypes = {
    router: routerShape.isRequired,
  };

  componentWillMount() {
    getAccessToken().then(token => {
      // TODO: Retrieve last environment
      let nextPath = "/default/default";
      if (token === null) {
        nextPath = "/login";
      }
      this.props.router.push(nextPath);
    });
  }

  render() {
    return null;
  }
}

export default RootRedirectPage;
