import React from "react";
import PropTypes from "prop-types";

import IconButton from "material-ui/IconButton";
import Button from "material-ui/ButtonBase";
import { withStyles } from "material-ui/styles";
import MenuIcon from "material-ui-icons/Menu";

import OrganizationIcon from "./OrganizationIcon";
import NamespaceSelector from "./NamespaceSelector";
import logo from "../assets/logo/wordmark/green.svg";

const styles = theme => ({
  header: {
    height: 172,
    backgroundColor: theme.palette.primary.dark,
  },
  row: {
    display: "flex",
    flexWrap: "wrap",
    justifyContent: "space-between",
  },
  logo: {
    height: 24,
    margin: "12px 12px 0 0",
  },
  selectorButton: {
    width: "100%",
    margin: "8px 0 -8px 0",
    padding: "8px 16px 8px 16px",
    display: "block",
    textAlign: "left",
  },
  selector: {
    width: "100%",
  },
  orgIcon: {
    margin: "24px 0 0 16px",
  },
  hamburgerButton: {
    color: theme.palette.primary.contrastText,
  },
});

class DrawerHeader extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    onToggle: PropTypes.func.isRequired,
  };

  render() {
    const { onToggle, classes } = this.props;

    return (
      <div className={classes.header}>
        <div className={classes.row}>
          <IconButton onClick={onToggle}>
            <MenuIcon />
          </IconButton>
          <img alt="sensu" src={logo} className={classes.logo} />
        </div>
        <div className={classes.row}>
          {/* TODO update with global variables or whatever when we get them */}
          <div className={classes.orgIcon}>
            <OrganizationIcon icon="HalfHeart" iconColor="#FA8072" size={36} />
          </div>
        </div>
        <div className={classes.row}>
          <Button aria-owns="test" className={classes.selectorButton}>
            <NamespaceSelector className={classes.selector} />
          </Button>
        </div>
      </div>
    );
  }
}
export default withStyles(styles)(DrawerHeader);
