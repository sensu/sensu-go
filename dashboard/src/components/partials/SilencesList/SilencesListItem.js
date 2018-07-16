import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Checkbox from "@material-ui/core/Checkbox";
import Disclosure from "@material-ui/icons/MoreVert";
import IconButton from "@material-ui/core/IconButton";
import Menu from "@material-ui/core/Menu";
import MenuController from "/components/controller/MenuController";
import MenuItem from "@material-ui/core/MenuItem";
import RelativeDate from "/components/RelativeDate";
import ResourceDetails from "/components/partials/ResourceDetails";
import RootRef from "@material-ui/core/RootRef";
import TableCell from "@material-ui/core/TableCell";
import TableOverflowCell from "/components/partials/TableOverflowCell";
import TableSelectableRow from "/components/partials/TableSelectableRow";

class SilencesListItem extends React.Component {
  static propTypes = {
    silence: PropTypes.object.isRequired,
    selected: PropTypes.bool.isRequired,
    onClickSelect: PropTypes.func.isRequired,
    onClickDelete: PropTypes.func.isRequired,
  };

  static fragments = {
    silence: gql`
      fragment SilencesListItem_silence on Silenced {
        storeId
        expireOnResolve
        expires
      }
    `,
  };

  renderExpiryCondition = () => {
    const { expires, expireOnResolve } = this.props.silence;
    if (expires && expireOnResolve) {
      return (
        <React.Fragment>
          Expires when <strong>resolved</strong> or{" "}
          <strong>
            <RelativeDate dateTime={expires} />
          </strong>.
        </React.Fragment>
      );
    } else if (expireOnResolve) {
      return (
        <React.Fragment>
          Expires when <strong>resolved</strong>.
        </React.Fragment>
      );
    } else if (expires) {
      return (
        <React.Fragment>
          Expires{" "}
          <strong>
            <RelativeDate dateTime={expires} />
          </strong>.
        </React.Fragment>
      );
    }
    return "Does not expire.";
  };

  render() {
    const { silence, selected, onClickSelect, onClickDelete } = this.props;

    return (
      <TableSelectableRow selected={selected}>
        <TableCell padding="checkbox">
          <Checkbox
            checked={selected}
            onChange={() => onClickSelect(!selected)}
          />
        </TableCell>
        <TableOverflowCell>
          <ResourceDetails
            title={<strong>{silence.storeId}</strong>}
            details={this.renderExpiryCondition()}
          />
        </TableOverflowCell>
        <TableCell padding="checkbox">
          <MenuController
            renderMenu={({ anchorEl, close }) => (
              <Menu open onClose={close} anchorEl={anchorEl}>
                <MenuItem
                  onClick={() => {
                    onClickDelete();
                    this.closeMenu();
                  }}
                >
                  Delete
                </MenuItem>
              </Menu>
            )}
          >
            {({ open, ref }) => (
              <RootRef rootRef={ref}>
                <IconButton onClick={open}>
                  <Disclosure />
                </IconButton>
              </RootRef>
            )}
          </MenuController>
        </TableCell>
      </TableSelectableRow>
    );
  }
}

export default SilencesListItem;
