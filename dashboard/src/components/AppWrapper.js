import React from "react";
import PropTypes from "prop-types";
import { Query } from "react-apollo";
import gql from "graphql-tag";

import AppFrame from "/components/AppFrame";

class AppWrapper extends React.Component {
  static propTypes = {
    children: PropTypes.element,
    environment: PropTypes.string.isRequired,
    organization: PropTypes.string.isRequired,
  };

  static defaultProps = { children: null };

  static query = gql`
    query AppWrapperQuery($environment: String!, $organization: String!) {
      viewer {
        ...AppFrame_viewer
      }
      environment(organization: $organization, environment: $environment) {
        ...AppFrame_environment
      }
    }
    ${AppFrame.fragments.viewer}
    ${AppFrame.fragments.environment}
  `;

  render() {
    return (
      <Query
        query={AppWrapper.query}
        variables={{
          organization: this.props.organization,
          environment: this.props.environment,
        }}
      >
        {({ data: { viewer, environment } = {}, loading, error }) => {
          if (error) throw error;

          return (
            <AppFrame
              loading={loading}
              viewer={viewer}
              environment={environment}
            >
              {this.props.children}
            </AppFrame>
          );
        }}
      </Query>
    );
  }
}

export default AppWrapper;
