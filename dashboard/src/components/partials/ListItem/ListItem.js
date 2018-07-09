import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";

import Disclosure from "@material-ui/icons/MoreVert";
import Checkbox from "@material-ui/core/Checkbox";
import IconButton from "@material-ui/core/IconButton";
import RootRef from "@material-ui/core/RootRef";
import TableCell from "@material-ui/core/TableCell";
import TableRow from "@material-ui/core/TableRow";

import MenuController from "/components/controller/MenuController";

const styles = theme => ({
  root: {
    verticalAlign: "top",

    // hover
    // https://material.io/guidelines/components/data-tables.html#data-tables-interaction
    "&:hover": {
      backgroundColor: theme.palette.action.selected,
    },
  },
  // selected
  // https://material.io/guidelines/components/data-tables.html#data-tables-interaction
  selected: {
    backgroundColor: theme.palette.action.hover,
  },
  cell: {
    paddingTop: 10,
    paddingBottom: 10,
    width: "100%",
    maxWidth: 0,
  },
  flex: {
    display: "flex",
    flexDirection: "row",
  },
  item: {
    lineHeight: "26px",
    color: theme.palette.text.secondary,

    "& strong": {
      color: theme.palette.text.primary,
    },

    "& a, & a:hover, & a:visited": {
      color: "inherit",
    },

    "&, & *": {
      whiteSpace: "nowrap",
      textOverflow: "ellipsis",
      overflow: "hidden",
    },
  },
  title: {
    fontSize: 14,
  },
  details: {
    fontSize: 13,
  },
  icon: {
    marginRight: 24,
  },
});

class ListItem extends React.PureComponent {
  static propTypes = {
    icon: PropTypes.node,
    title: PropTypes.node,
    details: PropTypes.node,
    selected: PropTypes.bool.isRequired,
    onChangeSelected: PropTypes.func.isRequired,
    renderMenu: PropTypes.func,
    classes: PropTypes.object.isRequired,
  };

  static defaultProps = {
    icon: undefined,
    title: undefined,
    details: undefined,
    renderMenu: undefined,
  };

  state = { menuOpen: false };

  _menuAnchorRef = React.createRef();

  openMenu = () => {
    this.setState({ menuOpen: true });
  };
  closeMenu = () => {
    this.setState({ menuOpen: false });
  };

  render() {
    const {
      icon,
      title,
      details,
      classes,
      selected,
      onChangeSelected,
      renderMenu,
    } = this.props;

    return (
      <TableRow
        className={classnames(classes.root, { [classes.selected]: selected })}
      >
        <TableCell padding="checkbox">
          <Checkbox
            color="primary"
            checked={selected}
            onChange={event => onChangeSelected(event.target.checked)}
          />
        </TableCell>
        <TableCell className={classes.cell} padding="none">
          <div className={classes.flex}>
            {icon && <div className={classes.icon}>{icon}</div>}
            <div className={classes.item}>
              <div className={classes.title}>{title}</div>
              <div className={classes.details}>{details}</div>
            </div>
          </div>
        </TableCell>
        <TableCell padding="checkbox">
          <MenuController renderMenu={renderMenu}>
            {({ open, ref }) => (
              <RootRef rootRef={ref}>
                <IconButton onClick={open}>
                  <Disclosure />
                </IconButton>
              </RootRef>
            )}
          </MenuController>
        </TableCell>
      </TableRow>
    );
  }
}

export default withStyles(styles)(ListItem);
