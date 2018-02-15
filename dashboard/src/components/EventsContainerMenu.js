import React from "react";
import PropTypes from "prop-types";

import map from "lodash/map";
import { withStyles } from "material-ui/styles";
import Typography from "material-ui/Typography";
import Button from "material-ui/ButtonBase";
import Menu, { MenuItem } from "material-ui/Menu";
import { ListItemText } from "material-ui/List";

import arrowIcon from "material-ui-icons/ArrowDropDown";

const styles = {
  tableHeaderButton: {
    marginLeft: 16,
    display: "flex",
  },
  arrow: { marginTop: -4 },
  checkbox: { marginTop: -4, width: 24, height: 24 },
};

class EventsContainerMenu extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    // eslint-disable-next-line react/forbid-prop-types
    contents: PropTypes.array.isRequired,
    label: PropTypes.string.isRequired,
    DropdownArrow: PropTypes.func.isRequired,
  };

  static defaultProps = { DropdownArrow: arrowIcon };

  state = {
    anchorEl: null,
  };

  onClose = () => {
    this.setState({ anchorEl: null });
  };

  handleClick = event => {
    this.setState({ anchorEl: event.currentTarget });
  };

  render() {
    const { classes, label, contents, DropdownArrow } = this.props;
    const { anchorEl } = this.state;

    return (
      <span>
        <Button onClick={this.handleClick}>
          <span className={classes.tableHeaderButton}>
            <Typography type="button">{label}</Typography>
            <DropdownArrow className={classes.arrow} />
          </span>
        </Button>

        <Menu
          anchorEl={anchorEl}
          open={Boolean(anchorEl)}
          onClose={this.onClose}
          id={`events-container-menu-${label}`}
        >
          {map(contents, name => (
            <MenuItem
              className={classes.menuItem}
              key={label}
              onClick={this.redirect}
            >
              <ListItemText primary={name} />
            </MenuItem>
          ))}
        </Menu>
      </span>
    );
  }
}

export default withStyles(styles)(EventsContainerMenu);
