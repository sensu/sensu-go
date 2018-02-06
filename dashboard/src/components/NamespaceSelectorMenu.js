import React from "react";
import PropTypes from "prop-types";
import partition from "lodash/partition";
import { map } from "lodash";
import { compose } from "recompose";
import { withStyles } from "material-ui/styles";
import { createFragmentContainer, graphql } from "react-relay";

import Menu, { MenuItem } from "material-ui/Menu";
import { ListItemIcon, ListItemText } from "material-ui/List";
import Divider from "material-ui/Divider";
import OrganizationIcon from "./OrganizationIcon";
import EnvironmentIcon from "./EnvironmentIcon";
import { withNamespace, namespaceShape } from "./NamespaceLink";

const styles = () => ({
  menuItem: {},
  primary: {},
  icon: {},
  environmentIconContainer: {
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    width: 24,
    height: 24,
  },
});

class NamespaceSelectorMenu extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    currentNamespace: namespaceShape.isRequired,
    viewer: PropTypes.objectOf(PropTypes.any).isRequired,
  };

  render() {
    const { viewer, currentNamespace, classes, ...props } = this.props;
    const navigateTo = (organization, environment) => () =>
      window.location.assign(`/${organization}/${environment}/`);

    const partitionedOrganizations = partition(
      viewer.organizations,
      organization => organization.name === currentNamespace.organization,
    );
    const currentOrganization = partitionedOrganizations[0][0];
    const otherOrganizations = partitionedOrganizations[1];

    return (
      <Menu {...props}>
        {currentOrganization.environments.map(environment => (
          <MenuItem
            className={classes.menuItem}
            key={environment.name}
            onClick={navigateTo(currentOrganization.name, environment.name)}
          >
            <ListItemIcon className={classes.icon}>
              <div className={classes.environmentIconContainer}>
                <EnvironmentIcon
                  color="rgb(250, 128, 114)"
                  className={classes.environmentIcon}
                  size={10}
                />
              </div>
            </ListItemIcon>
            <ListItemText
              classes={{ primary: classes.primary }}
              inset
              primary={environment.name}
            />
          </MenuItem>
        ))}
        <Divider />
        {map(otherOrganizations, (organization, i) => [
          organization.environments.map(environment => (
            <MenuItem
              className={classes.menuItem}
              key={environment.name}
              onClick={navigateTo(organization.name, environment.name)}
            >
              <ListItemIcon className={classes.icon}>
                <OrganizationIcon iconColor="#8AB8D0" />
              </ListItemIcon>
              <ListItemText
                classes={{ primary: classes.primary }}
                inset
                primary={organization.name}
                secondary={environment.name}
              />
            </MenuItem>
          )),
          i + 1 < otherOrganizations.length ? <Divider /> : null,
        ])}
      </Menu>
    );
  }
}

export default createFragmentContainer(
  compose(withStyles(styles), withNamespace)(NamespaceSelectorMenu),
  graphql`
    fragment NamespaceSelectorMenu_viewer on Viewer {
      organizations {
        name
        environments {
          name
        }
      }
    }
  `,
);
