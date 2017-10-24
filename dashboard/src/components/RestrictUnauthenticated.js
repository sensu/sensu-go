import React from "react";
import PropTypes from "prop-types";
import { withRouter, routerShape } from "found";
import { getAccessToken } from "../utils/authentication";

class RestrictUnauthenticated extends React.Component {
  static propTypes = {
    children: PropTypes.node.isRequired,
    router: routerShape.isRequired,
  };

  // TODO: Have something emit updates when access token is updated / revoked?
  componentWillMount() {
    getAccessToken().then(token => {
      if (!token) {
        this.props.router.push("/login");
      }
    });
  }

  render() {
    return this.props.children;
  }
}

export default withRouter(RestrictUnauthenticated);
