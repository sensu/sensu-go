import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { withStyles } from "@material-ui/core/styles";
import { Link } from "react-router-dom";
import Menu from "@material-ui/core/Menu";
import MenuItem from "@material-ui/core/MenuItem";
import ListItemText from "@material-ui/core/ListItemText";
import ListItemIcon from "@material-ui/core/ListItemIcon";
import Divider from "@material-ui/core/Divider";
import NamespaceIcon from "/components/partials/NamespaceIcon";

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
        namespaces {
          name
          ...NamespaceIcon_namespace
        }
      }

      ${NamespaceIcon.fragments.namespace}
    `,
  };

  render() {
    const { viewer, classes, org, onClick, ...props } = this.props;

    if (!viewer) {
      return null;
    }

    const groupedNamespaces = viewer.namespaces
      .map(ns => ns.split("/", 1))
      .reduce((acc, ns) => {
        acc[ns[0]] = acc[ns[0]] || [];
        acc[ns[0]].append(ns.join("/"));
        return acc;
      }, {});

    return (
      <Menu {...props}>
        {Object.keys(groupedNamespaces).map((key, i) => {
          const namesapces = groupedNamespaces[key];

          return (
            <React.Fragment key={`prefix-${key}`}>
              {namesapces.map(namespace => (
                <Link to={`/${namespace}`} onClick={onClick}>
                  <MenuItem>
                    <ListItemIcon>
                      <NamespaceIcon namespace={namespace} size={24} />
                    </ListItemIcon>
                    <ListItemText inset primary={namespace} />
                  </MenuItem>
                </Link>
              ))}
              {i + 1 < groupedNamespaces.length && <Divider />}
            </React.Fragment>
          );
        })}
      </Menu>
    );
  }
}

export default withStyles(styles)(NamespaceSelectorMenu);
