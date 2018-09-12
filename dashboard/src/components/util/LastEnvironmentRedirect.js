import React from "react";
import PropTypes from "prop-types";
import { Redirect, withRouter } from "react-router-dom";
import { redirectKey } from "/constants/queryParams";
import gql from "graphql-tag";

import Query from "/components/util/Query";

const primaryQuery = gql`
  query LastEnvironmentQuery {
    lastEnvironment @client {
      organization
      environment
    }
  }
`;

const fallbackQuery = gql`
  query LastEnvironmentFallbackQuery {
    viewer {
      organizations {
        name
        environments {
          name
        }
      }
    }
  }
`;

class LastEnvironmentRedirect extends React.PureComponent {
  static propTypes = {
    // from withRouter HOC
    location: PropTypes.object.isRequired,
  };

  renderFallback = ({ data, aborted, loading }) => {
    if (loading || aborted) {
      return null;
    }

    if (data.viewer && data.viewer.organizations.length === 0) {
      return <Redirect to={"/default/default"} />;
    }

    const firstOrg = data.viewer.organizations[0];
    if (firstOrg.environments.length === 0) {
      return <Redirect to={`/${firstOrg.name}/default`} />;
    }

    const firstEnv = firstOrg.environments[0];
    return <Redirect to={`/${firstOrg.name}/${firstEnv.name}`} />;
  };

  renderRedirect = ({ data }) => {
    const { location } = this.props;

    // 1. if 'redirect-to' query parameter is present use given path.
    const queryParams = new URLSearchParams(location.search);
    const redirectQueryParam = queryParams.get(redirectKey);
    if (redirectQueryParam) {
      return <Redirect to={redirectQueryParam} />;
    }

    // 2. if the user's last environment was not recovered from localStorage
    // we fetch all the user's organizations and redirect to first result.
    const { lastEnvironment } = data;
    if (!lastEnvironment) {
      return <Query query={fallbackQuery}>{this.renderFallback}</Query>;
    }

    // 3. if the user's last environment is available, build path.
    const nextPath = [
      lastEnvironment.organization,
      lastEnvironment.environment,
    ].join("/");
    return <Redirect to={nextPath} />;
  };

  render() {
    return <Query query={primaryQuery}>{this.renderRedirect}</Query>;
  }
}

export default withRouter(LastEnvironmentRedirect);
