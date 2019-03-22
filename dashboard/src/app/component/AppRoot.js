import React from "react";
import PropTypes from "prop-types";
import { compose } from "recompose";
import { Provider } from "react-redux";
import { ApolloProvider } from "react-apollo";
import { withStyles } from "@material-ui/core/styles";
import { Switch, Route, withRouter } from "react-router-dom";

import AppThemeProvider from "/lib/component/base/AppThemeProvider";
import ResetStyles from "/lib/component/base/ResetStyles";
import ThemeStyles from "/lib/component/base/ThemeStyles";

import GlobalAlert from "/lib/component/util/GlobalAlert";
import AuthenticatedRoute from "/lib/component/util/AuthenticatedRoute";
import UnauthenticatedRoute from "/lib/component/util/UnauthenticatedRoute";
import AuthInvalidRoute from "/lib/component/util/AuthInvalidRoute";
import DefaultRedirect from "/lib/component/util/DefaultRedirect";
import LastNamespaceRedirect from "/lib/component/util/LastNamespaceRedirect";
import SigninRedirect from "/lib/component/util/SigninRedirect";
import { Provider as RelocationProvider } from "/lib/component/relocation/Relocation";

import EnvironmentView from "/app/component/views/EnvironmentView";
import SignInView from "/app/component/views/SignInView";
import NotFoundView from "/app/component/views/NotFoundView";

import AuthInvalidDialog from "/app/component/partial/AuthInvalidDialog";

class AppRoot extends React.PureComponent {
  static propTypes = {
    apolloClient: PropTypes.object.isRequired,
    reduxStore: PropTypes.object.isRequired,
  };

  render() {
    const { reduxStore, apolloClient } = this.props;

    return (
      <RelocationProvider>
        <Provider store={reduxStore}>
          <ApolloProvider client={apolloClient}>
            <AppThemeProvider>
              <Switch>
                <Route exact path="/" component={DefaultRedirect} />
                <UnauthenticatedRoute
                  exact
                  path="/signin"
                  component={SignInView}
                  fallbackComponent={LastNamespaceRedirect}
                />
                <AuthenticatedRoute
                  path="/:namespace"
                  component={EnvironmentView}
                  fallbackComponent={SigninRedirect}
                />
                <Route component={NotFoundView} />
              </Switch>
              <Switch>
                <UnauthenticatedRoute exact path="/signin" />
                <AuthInvalidRoute component={AuthInvalidDialog} />
              </Switch>
              <ResetStyles />
              <ThemeStyles />
              <GlobalAlert />
            </AppThemeProvider>
          </ApolloProvider>
        </Provider>
      </RelocationProvider>
    );
  }
}

// AppRoot is composed with `withRouter` here to prevent "Update Blocking"
// see: reacttraining.com/react-router/web/guides/dealing-with-update-blocking
export default compose(
  withStyles(AppRoot.styles),
  withRouter,
)(AppRoot);
