import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
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

  static fragments = {
    environment: gql`
      fragment EnvironmentLabel_environment on Environment {
        name
        colourId
        organization {
          name
          iconId
        }
      }
    `,
  };

  render() {
    const { environment, ...props } = this.props;
    const { organization } = environment;

    return (
      <Label
        organization={organization.name}
        organizationIcon={organization.iconId}
        environment={environment.name}
        environmentColour={environment.colourId}
        {...props}
      />
    );
  }
}

export default EnvironmentLabel;
