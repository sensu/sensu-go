import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Button from "@material-ui/core/ButtonBase";
import Checkbox from "@material-ui/core/Checkbox";
import Disclosure from "@material-ui/icons/MoreVert";
import Menu from "@material-ui/core/Menu";
import MenuItem from "@material-ui/core/MenuItem";
import TableCell from "@material-ui/core/TableCell";
import TableRow from "@material-ui/core/TableRow";

class CheckListItem extends React.Component {
  static propTypes = {
    check: PropTypes.object.isRequired,
    selected: PropTypes.bool.isRequired,
    setSelected: PropTypes.func.isRequired,
    onClickSilence: PropTypes.func.isRequired,
  };

  static fragments = {
    check: gql`
      fragment ChecksListItem_check on CheckConfig {
        name
        command
        subscriptions
        interval
      }
    `,
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
    const { check, selected, setSelected, onClickSilence } = this.props;

    return (
      <TableRow>
        <TableCell padding="checkbox">
          <Checkbox
            checked={selected}
            onChange={event => setSelected(event.target.checked)}
          />
        </TableCell>
        <TableCell style={{ width: "100%" }}>
          {check.name}
          <br />
          {check.command}
        </TableCell>
        <TableCell>
          <div ref={this._menuAnchorRef}>
            <Button onClick={this.openMenu}>
              <Disclosure />
            </Button>
          </div>
          <Menu
            open={this.state.menuOpen}
            onClose={this.closeMenu}
            anchorEl={this._menuAnchorRef.current}
          >
            <MenuItem
              onClick={() => {
                onClickSilence();
                this.closeMenu();
              }}
            >
              Silence
            </MenuItem>
          </Menu>
        </TableCell>
      </TableRow>
    );
  }
}

export default CheckListItem;
