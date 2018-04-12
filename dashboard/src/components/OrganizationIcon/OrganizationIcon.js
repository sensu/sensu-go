import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import Icon from "./OrganizationIconBase";

class OrganizationIcon extends React.Component {
  static propTypes = {
    organization: PropTypes.shape({
      iconId: PropTypes.string,
    }).isRequired,
  };

  static fragments = {
    organization: gql`
      fragment OrganizationIcon_organization on Organization {
        iconId
      }
    `,
  };

  render() {
    const { organization, ...props } = this.props;
    return <Icon icon={organization.iconId} {...props} />;
  }
}

export default OrganizationIcon;
