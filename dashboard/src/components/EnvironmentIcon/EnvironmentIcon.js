import React from "react";
import PropTypes from "prop-types";
import { createFragmentContainer, graphql } from "react-relay";
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

  render() {
    const { environment: { organization, colourId }, ...props } = this.props;
    return (
      <Icon
        organizationIcon={organization.iconId}
        environmentColour={colourId}
        {...props}
      />
    );
  }
}

export default createFragmentContainer(
  EnvironmentIcon,
  graphql`
    fragment EnvironmentIcon_environment on Environment {
      colourId
      organization {
        iconId
      }
    }
  `,
);
