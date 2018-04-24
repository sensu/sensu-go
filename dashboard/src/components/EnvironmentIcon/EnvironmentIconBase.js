import React from "react";
import PropTypes from "prop-types";

import { OrganizationIconBase as Icon } from "/components/OrganizationIcon";
import { EnvironmentSymbolBase as Symbol } from "/components/EnvironmentSymbol";

class EnvironmentIcon extends React.Component {
  static propTypes = {
    organizationIcon: PropTypes.string.isRequired,
    environmentColour: PropTypes.string.isRequired,
    size: PropTypes.number,
  };

  static defaultProps = {
    size: 24.0,
    disableEnvironmentIdicator: false,
  };

  render() {
    const { size, organizationIcon, environmentColour, ...props } = this.props;

    return (
      <Icon icon={organizationIcon} size={size} {...props}>
        <Symbol colour={environmentColour} size={size / 3.0} />
      </Icon>
    );
  }
}

export default EnvironmentIcon;
