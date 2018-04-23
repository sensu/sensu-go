import React from "react";
import { hoistStatics } from "recompose";

import * as tokenStore from "/utils/authentication/tokens";

const withAuthTokens = hoistStatics(Component => {
  class WithAuthTokens extends React.PureComponent {
    state = { authTokens: tokenStore.get() };

    componentWillMount() {
      tokenStore.subscribe(this._handleTokens);
    }

    componentWillUnmount() {
      tokenStore.unsubscribe(this._handleTokens);
    }

    _handleTokens = authTokens => this.setState({ authTokens });

    render() {
      return <Component {...this.props} authTokens={this.state.authTokens} />;
    }
  }

  return WithAuthTokens;
});

export default withAuthTokens;
