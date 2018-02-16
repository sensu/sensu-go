import React from "react";
import PropTypes from "prop-types";

import { withStyles } from "material-ui/styles";
import Typography from "material-ui/Typography";
import Button from "material-ui/ButtonBase";
import Menu, { MenuItem } from "material-ui/Menu";
import { ListItemText } from "material-ui/List";

import arrowIcon from "material-ui-icons/ArrowDropDown";

import EventStatus from "./EventStatus";

const styles = {
  tableHeaderButton: {
    marginLeft: 16,
    display: "flex",
  },
  arrow: { marginTop: -4 },
  checkbox: { marginTop: -4, width: 24, height: 24 },
  humanStatus: { marginLeft: 5 },
};

class EventsContainerMenu extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    // eslint-disable-next-line react/forbid-prop-types
    contents: PropTypes.array.isRequired,
    label: PropTypes.string.isRequired,
    DropdownArrow: PropTypes.func.isRequired,
    icons: PropTypes.bool,
    onSelectValue: PropTypes.func.isRequired,
  };

  static defaultProps = {
    DropdownArrow: arrowIcon,
    icons: false,
  };

  state = {
    anchorEl: null,
    selectValue: null,
  };

  onClose = () => {
    this.setState({ anchorEl: null });
  };

  handleClick = event => {
    this.setState({ anchorEl: event.currentTarget });
  };

  selectValue = value => () => {
    this.props.onSelectValue(value);
    this.setState({ anchorEl: null });
  };

  render() {
    const { classes, label, icons, contents, DropdownArrow } = this.props;
    const { anchorEl } = this.state;

    let items = {};
    if (!icons) {
      items = contents.map((name, i) => (
        <MenuItem
          className={classes.menuItem}
          // eslint-disable-next-line react/no-array-index-key
          key={`${label}-${i}`}
          onClick={this.selectValue(name)}
        >
          <ListItemText primary={name} />
        </MenuItem>
      ));
    } else {
      const humanStatuses = ["Passing", "Warning", "Critical", "Unknown"];
      // TODO this code is probably reusable/make new component
      items = contents.map((status, i) => (
        <MenuItem
          className={classes.menuItem}
          // eslint-disable-next-line react/no-array-index-key
          key={`${label}-${i}`}
          onClick={this.selectValue(status)}
        >
          <EventStatus status={status} />
          <span className={classes.humanStatus}>
            {status > 2 ? "Unknown" : humanStatuses[status]}
          </span>
        </MenuItem>
      ));
    }

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
          {items}
        </Menu>
      </span>
    );
  }
}

export default withStyles(styles)(EventsContainerMenu);
