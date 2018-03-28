import React from "react";
import PropTypes from "prop-types";
import { createFragmentContainer, graphql } from "react-relay";
import Label from "./EnvironmentLabelBase";

class EnvironmentLabel extends React.Component {
  static propTypes = {
    environment: PropTypes.shape({
      name: PropTypes.string,
      colourId: PropTypes.string,
      organization: PropTypes.shape({
        name: PropTypes.string,
        iconId: PropTypes.string,
      }).isRequired,
    }).isRequired,
  };

  render() {
    const { environment } = this.props;
    const { organization } = environment;

    return (
      <Label
        organization={organization.name}
        environment={environment.name}
        icon={organization.iconId}
        iconColour={environment.colourId}
      />
    );
  }
}

export default createFragmentContainer(
  EnvironmentLabel,
  graphql`
    fragment EnvironmentLabel_environment on Environment {
      name
      colourId
      organization {
        name
        iconId
      }
    }
  `,
);
