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
import NamespaceIcon from "/app/component/partial/NamespaceIcon";

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
    const { viewer, classes, onClick, ...props } = this.props;

    if (!viewer) {
      return null;
    }

    const groups = viewer.namespaces.reduce((acc, ns) => {
      const [key] = ns.name.split("-", 1);

      acc[key] = acc[key] || [];
      acc[key].push(ns);

      return acc;
    }, {});

    return (
      <Menu {...props}>
        {Object.keys(groups).map((key, i) => {
          const namesapces = groups[key];

          return (
            <React.Fragment key={`prefix-${key}`}>
              {namesapces.map(namespace => (
                <Link
                  key={namespace.name}
                  to={`/${namespace.name}`}
                  onClick={onClick}
                >
                  <MenuItem>
                    <ListItemIcon>
                      <NamespaceIcon namespace={namespace} size={24} />
                    </ListItemIcon>
                    <ListItemText inset primary={namespace.name} />
                  </MenuItem>
                </Link>
              ))}

              {i + 1 < groups.length && <Divider />}
            </React.Fragment>
          );
        })}
      </Menu>
    );
  }
}

export default withStyles(styles)(NamespaceSelectorMenu);
