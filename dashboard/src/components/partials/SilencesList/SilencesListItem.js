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

class SilencesListItem extends React.Component {
  static propTypes = {
    silence: PropTypes.object.isRequired,
    selected: PropTypes.bool.isRequired,
    setSelected: PropTypes.func.isRequired,
    onClickDelete: PropTypes.func.isRequired,
  };

  static fragments = {
    silence: gql`
      fragment SilencesListItem_silence on Silenced {
        storeId
        creator {
          username
        }
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
    const { silence, selected, setSelected, onClickDelete } = this.props;

    return (
      <TableRow>
        <TableCell padding="checkbox">
          <Checkbox
            checked={selected}
            onChange={event => setSelected(event.target.checked)}
          />
        </TableCell>
        <TableCell style={{ width: "100%" }}>
          <b>{silence.storeId}</b>
          <br />
          <sub>
            {silence.expiresOnResolve ? "Expires on resolve" : "Never expires"}
          </sub>
        </TableCell>
        <TableCell>{silence.creator.username}</TableCell>
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
                onClickDelete();
                this.closeMenu();
              }}
            >
              Delete
            </MenuItem>
          </Menu>
        </TableCell>
      </TableRow>
    );
  }
}

export default SilencesListItem;
