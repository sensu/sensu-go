import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import partition from "lodash/partition";
import { withStyles } from "@material-ui/core/styles";
import { Link } from "react-router-dom";
import Menu from "@material-ui/core/Menu";
import MenuItem from "@material-ui/core/MenuItem";
import ListItemText from "@material-ui/core/ListItemText";
import ListItemIcon from "@material-ui/core/ListItemIcon";
import Divider from "@material-ui/core/Divider";
import OrganizationIcon from "/components/OrganizationIcon";
import EnvironmentSymbol from "/components/EnvironmentSymbol";

const styles = () => ({
  envIcon: {
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    width: 24,
    height: 24,
  },
});

class NamespaceSelectorMenu extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    org: PropTypes.string.isRequired,
    onClick: PropTypes.func.isRequired,
    viewer: PropTypes.object,
  };

  static defaultProps = {
    viewer: null,
  };

  static fragments = {
    viewer: gql`
      fragment NamespaceSelectorMenu_viewer on Viewer {
        organizations {
          name
          ...OrganizationIcon_organization
          environments {
            name
            ...EnvironmentSymbol_environment
          }
        }
      }

      ${OrganizationIcon.fragments.organization}
      ${EnvironmentSymbol.fragments.environment}
    `,
  };

  render() {
    const { viewer, classes, org, onClick, ...props } = this.props;

    if (!viewer) {
      return null;
    }

    const [[currentOrganization], otherOrganizations] = partition(
      viewer.organizations,
      organization => organization.name === org,
    );

    return (
      <Menu {...props}>
        {currentOrganization &&
          currentOrganization.environments.map(environment => (
            <Link
              to={`/${currentOrganization.name}/${environment.name}`}
              onClick={onClick}
              key={environment.name}
            >
              <MenuItem>
                <ListItemIcon>
                  <div className={classes.envIcon}>
                    <EnvironmentSymbol environment={environment} size={12} />
                  </div>
                </ListItemIcon>
                <ListItemText inset primary={environment.name} />
              </MenuItem>
            </Link>
          ))}
        <Divider />
        {otherOrganizations.map((organization, i) => [
          organization.environments.map(environment => (
            <Link
              to={`/${organization.name}/${environment.name}`}
              key={`${organization.name}/${environment.name}`}
              onClick={onClick}
            >
              <MenuItem>
                <ListItemIcon>
                  <OrganizationIcon organization={organization} size={24}>
                    <EnvironmentSymbol
                      environment={environment}
                      size={24 / 3}
                    />
                  </OrganizationIcon>
                </ListItemIcon>
                <ListItemText
                  inset
                  primary={organization.name}
                  secondary={environment.name}
                />
              </MenuItem>
            </Link>
          )),
          i + 1 < otherOrganizations.length ? <Divider /> : null,
        ])}
      </Menu>
    );
  }
}

export default withStyles(styles)(NamespaceSelectorMenu);
