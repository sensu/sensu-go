import React from "react";
import PropTypes from "prop-types";
import { compose } from "recompose";
import { Provider } from "react-redux";
import { ApolloProvider } from "react-apollo";
import { withStyles } from "material-ui/styles";
import { Switch, Route, withRouter } from "react-router-dom";

import AppThemeProvider from "./AppThemeProvider";
import ResetStyles from "./ResetStyles";
import ThemeStyles from "./ThemeStyles";

import AuthenticatedRoute from "./util/AuthenticatedRoute";
import UnauthenticatedRoute from "./util/UnauthenticatedRoute";

import DefaultRedirect from "./util/DefaultRedirect";

import EnvironmentView from "./views/EnvironmentView";
import LoginView from "./views/LoginView";
import NotFoundView from "./views/NotFoundView";

class AppRoot extends React.PureComponent {
  static propTypes = {
    reduxStore: PropTypes.object.isRequired,
    apolloClient: PropTypes.object.isRequired,
  };

  static defaultProps = { children: null };

  render() {
    const { reduxStore, apolloClient } = this.props;

    return (
      <Provider store={reduxStore}>
        <ApolloProvider client={apolloClient}>
          <AppThemeProvider>
            <Switch>
              <Route exact path="/" component={DefaultRedirect} />
              <UnauthenticatedRoute
                exact
                path="/login"
                component={LoginView}
                fallbackComponent={DefaultRedirect}
              />
              <AuthenticatedRoute
                path="/:organization/:environment"
                component={EnvironmentView}
                fallbackComponent={DefaultRedirect}
              />
              <Route component={NotFoundView} />
            </Switch>
            <ResetStyles />
            <ThemeStyles />
          </AppThemeProvider>
        </ApolloProvider>
      </Provider>
    );
  }
}

// AppRoot is composed with `withRouter` here to prevent "Update Blocking"
// see: reacttraining.com/react-router/web/guides/dealing-with-update-blocking
export default compose(withStyles(AppRoot.styles), withRouter)(AppRoot);
