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

import GlobalAlert from "/components/util/GlobalAlert";
import AuthenticatedRoute from "/components/util/AuthenticatedRoute";
import UnauthenticatedRoute from "/components/util/UnauthenticatedRoute";
import AuthInvalidRoute from "/components/util/AuthInvalidRoute";
import DefaultRedirect from "/components/util/DefaultRedirect";
import LastNamespaceRedirect from "/components/util/LastNamespaceRedirect";
import SigninRedirect from "/components/util/SigninRedirect";
import { Provider as RelocationProvider } from "/components/relocation/Relocation";

import EnvironmentView from "/components/views/EnvironmentView";
import SignInView from "/components/views/SignInView";
import NotFoundView from "/components/views/NotFoundView";

import AuthInvalidDialog from "/components/partials/AuthInvalidDialog";

class AppRoot extends React.PureComponent {
  static propTypes = {
    reduxStore: PropTypes.object.isRequired,
    apolloClient: PropTypes.object.isRequired,
  };

  static defaultProps = { children: null };

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
export default compose(withStyles(AppRoot.styles), withRouter)(AppRoot);
