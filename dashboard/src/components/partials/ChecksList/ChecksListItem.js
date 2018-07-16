import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Checkbox from "@material-ui/core/Checkbox";
import IconButton from "@material-ui/core/IconButton";
import Menu from "@material-ui/core/Menu";
import MenuItem from "@material-ui/core/MenuItem";
import MoreVert from "@material-ui/icons/MoreVert";
import RootRef from "@material-ui/core/RootRef";
import TableCell from "@material-ui/core/TableCell";

import MenuController from "/components/controller/MenuController";

import ResourceDetails from "/components/partials/ResourceDetails";
import TableOverflowCell from "/components/partials/TableOverflowCell";
import TableSelectableRow from "/components/partials/TableSelectableRow";

import Code from "/components/Code";

class CheckListItem extends React.Component {
  static propTypes = {
    check: PropTypes.object.isRequired,
    selected: PropTypes.bool.isRequired,
    onChangeSelected: PropTypes.func.isRequired,
    onClickExecute: PropTypes.func.isRequired,
    onClickSilence: PropTypes.func.isRequired,
  };

  static fragments = {
    check: gql`
      fragment ChecksListItem_check on CheckConfig {
        name
        command
        source
        subscriptions
        interval
      }
    `,
  };

  render() {
    const {
      check,
      selected,
      onChangeSelected,
      onClickSilence,
      onClickExecute,
    } = this.props;

    return (
      <TableSelectableRow selected={selected}>
        <TableCell padding="checkbox">
          <Checkbox
            color="primary"
            checked={selected}
            onChange={e => onChangeSelected(e.target.checked)}
          />
        </TableCell>
        <TableOverflowCell>
          <ResourceDetails
            title={<strong>{check.name}</strong>}
            details={
              <React.Fragment>
                <Code>{check.command}</Code>
                <br />
                Executed every{" "}
                <strong>
                  {check.interval} {check.interval === 1 ? "second" : "seconds"}
                </strong>{" "}
                by{" "}
                <strong>
                  {check.subscriptions.length}{" "}
                  {check.subscriptions.length === 1
                    ? "subscription"
                    : "subscriptions"}
                </strong>.
              </React.Fragment>
            }
          />
        </TableOverflowCell>

        <TableCell padding="checkbox">
          <MenuController
            renderMenu={({ anchorEl, close }) => (
              <Menu open onClose={close} anchorEl={anchorEl}>
                <MenuItem
                  onClick={() => {
                    onClickExecute();
                    close();
                  }}
                >
                  Execute
                </MenuItem>
                <MenuItem
                  onClick={() => {
                    onClickSilence();
                    close();
                  }}
                >
                  Silence
                </MenuItem>
              </Menu>
            )}
          >
            {({ open, ref }) => (
              <RootRef rootRef={ref}>
                <IconButton onClick={open}>
                  <MoreVert />
                </IconButton>
              </RootRef>
            )}
          </MenuController>
        </TableCell>
      </TableSelectableRow>
    );
  }
}

export default CheckListItem;
