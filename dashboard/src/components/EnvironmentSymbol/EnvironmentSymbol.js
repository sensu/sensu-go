import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import Symbol from "./EnvironmentSymbolBase";

class EnvironmentSymbol extends React.Component {
  static propTypes = {
    environment: PropTypes.shape({
      colourId: PropTypes.string,
    }).isRequired,
  };

  static fragments = {
    environment: gql`
      fragment EnvironmentSymbol_environment on Environment {
        colourId
      }
    `,
  };

  render() {
    const { environment, ...props } = this.props;
    return <Symbol colour={environment.colourId} {...props} />;
  }
}

export default EnvironmentSymbol;
