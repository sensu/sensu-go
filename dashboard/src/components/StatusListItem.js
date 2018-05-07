import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "material-ui/styles";
import Button from "material-ui/ButtonBase";
import Checkbox from "material-ui/Checkbox";
import Disclosure from "material-ui-icons/MoreVert";

import CheckStatusIcon from "/components/CheckStatusIcon";
import { TableListItem } from "/components/TableList";

const styles = theme => ({
  root: {
    color: theme.palette.text.secondary,
    "& strong": {
      fontWeight: "normal",
      color: theme.palette.text.primary,
    },
  },
  checkbox: {
    display: "inline-block",
    verticalAlign: "top",
    marginLeft: -12,
  },
  status: {
    display: "inline-block",
    verticalAlign: "top",
    padding: "14px 0",
  },
  disclosure: {
    color: theme.palette.action.active,
    marginLeft: 12,
    paddingTop: 14,
  },
  content: {
    width: "calc(100% - 96px)",
    flex: 1,
    padding: 14,
  },
  title: {
    marginBottom: 4,
  },
  details: {
    width: "100%",
    fontSize: "0.8125rem",
  },
});

class StatusListItem extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    selected: PropTypes.bool,
    deleted: PropTypes.bool,
    onClickSelect: PropTypes.func,
    renderMenu: PropTypes.func,
    status: PropTypes.number,
    title: PropTypes.node,
    children: PropTypes.node,
  };

  static defaultProps = {
    renderMenu: undefined,
    selected: undefined,
    deleted: undefined,
    onClickSelect: undefined,
    status: undefined,
    title: null,
    children: null,
  };

  state = { menuOpen: false };

  _menuAnchorRef = null;
  _handleMenuAnchorRef = ref => {
    this._menuAnchorRef = ref;
  };

  openMenu = () => {
    this.setState({ menuOpen: true });
  };

  closeMenu = () => {
    this.setState({ menuOpen: false });
  };

  render() {
    const {
      status,
      selected,
      deleted,
      classes,
      onClickSelect,
      renderMenu,
      title,
      children,
    } = this.props;
    const { menuOpen } = this.state;

    return (
      <TableListItem className={classes.root} selected={selected}>
        <div className={classes.checkbox}>
          <Checkbox
            color="primary"
            onChange={onClickSelect}
            checked={selected}
            disabled={deleted}
          />
        </div>
        <div className={classes.status}>
          <CheckStatusIcon statusCode={status} />
        </div>
        <div className={classes.content}>
          <div className={classes.title}>{title}</div>
          {children && <div className={classes.details}>{children}</div>}
        </div>
        {renderMenu && (
          <div className={classes.disclosure} ref={this._handleMenuAnchorRef}>
            {!deleted && (
              <Button onClick={this.openMenu}>
                <Disclosure />
              </Button>
            )}
            {!deleted &&
              renderMenu({
                open: menuOpen,
                onClose: this.closeMenu,
                anchorEl: this._menuAnchorRef,
              })}
          </div>
        )}
      </TableListItem>
    );
  }
}

export default withStyles(styles)(StatusListItem);
