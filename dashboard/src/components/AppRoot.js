import React from "react";
import PropTypes from "prop-types";
import { compose } from "recompose";
import { Provider } from "react-redux";
import { ApolloProvider } from "react-apollo";
import { withStyles } from "@material-ui/core/styles";
import { Switch, Route, withRouter } from "react-router-dom";

import AppThemeProvider from "/components/AppThemeProvider";
import ResetStyles from "/components/ResetStyles";
import ThemeStyles from "/components/ThemeStyles";

import AuthenticatedRoute from "/components/util/AuthenticatedRoute";
import UnauthenticatedRoute from "/components/util/UnauthenticatedRoute";

import DefaultRedirect from "/components/util/DefaultRedirect";

import EnvironmentView from "/components/views/EnvironmentView";
import SignInView from "/components/views/SignInView";
import NotFoundView from "/components/views/NotFoundView";

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
                path="/signin"
                component={SignInView}
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
