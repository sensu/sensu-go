import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import Icon from "./EnvironmentIconBase";

class EnvironmentIcon extends React.Component {
  static propTypes = {
    environment: PropTypes.shape({
      colorId: PropTypes.string,
      organization: PropTypes.shape({
        iconId: PropTypes.string,
      }),
    }).isRequired,
  };

  static fragments = {
    environment: gql`
      fragment EnvironmentIcon_environment on Environment {
        colourId
        organization {
          iconId
        }
      }
    `,
  };

  render() {
    const {
      environment: { organization, colourId },
      ...props
    } = this.props;
    return (
      <Icon
        organizationIcon={organization.iconId}
        environmentColour={colourId}
        {...props}
      />
    );
  }
}

export default EnvironmentIcon;
