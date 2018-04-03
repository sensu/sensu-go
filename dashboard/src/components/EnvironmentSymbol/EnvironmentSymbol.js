import React from "react";
import PropTypes from "prop-types";
import { createFragmentContainer, graphql } from "react-relay";
import Symbol from "./EnvironmentSymbolBase";

class EnvironmentSymbol extends React.Component {
  static propTypes = {
    environment: PropTypes.shape({
      colourId: PropTypes.string,
    }).isRequired,
  };

  render() {
    const { environment, ...props } = this.props;
    return <Symbol colour={environment.colourId} {...props} />;
  }
}

export default createFragmentContainer(
  EnvironmentSymbol,
  graphql`
    fragment EnvironmentSymbol_environment on Environment {
      colourId
    }
  `,
);
