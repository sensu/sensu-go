import React from "react";
import PropTypes from "prop-types";
import { createFragmentContainer, graphql } from "react-relay";
import Icon from "./OrganizationIconBase";

class OrganizationIcon extends React.Component {
  static propTypes = {
    organization: PropTypes.shape({
      iconId: PropTypes.string,
    }).isRequired,
  };

  render() {
    const { organization, ...props } = this.props;
    return <Icon icon={organization.iconId} {...props} />;
  }
}

export default createFragmentContainer(
  OrganizationIcon,
  graphql`
    fragment OrganizationIcon_organization on Organization {
      iconId
    }
  `,
);
