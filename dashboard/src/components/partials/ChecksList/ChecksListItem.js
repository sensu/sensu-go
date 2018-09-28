import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Checkbox from "@material-ui/core/Checkbox";
import ConfirmDelete from "/components/partials/ConfirmDelete";
import IconButton from "@material-ui/core/IconButton";
import Menu from "@material-ui/core/Menu";
import MenuController from "/components/controller/MenuController";
import MenuItem from "@material-ui/core/MenuItem";
import MoreVert from "@material-ui/icons/MoreVert";
import NamespaceLink from "/components/util/NamespaceLink";
import ResourceDetails from "/components/partials/ResourceDetails";
import RootRef from "@material-ui/core/RootRef";
import SilenceIcon from "/icons/Silence";
import TableCell from "@material-ui/core/TableCell";
import TableOverflowCell from "/components/partials/TableOverflowCell";
import TableSelectableRow from "/components/partials/TableSelectableRow";
import Code from "/components/Code";
import CheckSchedule from "./CheckSchedule";

class CheckListItem extends React.Component {
  static propTypes = {
    check: PropTypes.object.isRequired,
    selected: PropTypes.bool.isRequired,
    onChangeSelected: PropTypes.func.isRequired,
    onClickClearSilences: PropTypes.func.isRequired,
    onClickDelete: PropTypes.func.isRequired,
    onClickExecute: PropTypes.func.isRequired,
    onClickSetPublish: PropTypes.func.isRequired,
    onClickSilence: PropTypes.func.isRequired,
  };

  static fragments = {
    check: gql`
      fragment ChecksListItem_check on CheckConfig {
        name
        command
        isSilenced
        namespace {
          organization
          environment
        }
        ...CheckSchedule_check
      }

      ${CheckSchedule.fragments.check}
    `,
  };

  render() {
    const {
      check,
      selected,
      onChangeSelected,
      onClickClearSilences,
      onClickDelete,
      onClickExecute,
      onClickSetPublish,
      onClickSilence,
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
            title={
              <NamespaceLink
                namespace={check.namespace}
                to={`/checks/${check.name}`}
              >
                <strong>{check.name} </strong>
                {check.isSilenced && (
                  <SilenceIcon
                    fontSize="inherit"
                    style={{ verticalAlign: "text-top" }}
                  />
                )}
              </NamespaceLink>
            }
            details={
              <React.Fragment>
                <Code>{check.command}</Code>
                <br />
                <CheckSchedule check={check} />
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
                {!check.isSilenced && (
                  <MenuItem
                    onClick={() => {
                      onClickSilence();
                      close();
                    }}
                  >
                    Silence
                  </MenuItem>
                )}
                {check.isSilenced && (
                  <MenuItem
                    onClick={() => {
                      onClickClearSilences();
                      close();
                    }}
                  >
                    Unsilence
                  </MenuItem>
                )}
                {!check.publish && (
                  <MenuItem
                    onClick={() => {
                      onClickSetPublish(true);
                      close();
                    }}
                  >
                    Publish
                  </MenuItem>
                )}
                {check.publish && (
                  <MenuItem
                    onClick={() => {
                      onClickSetPublish(false);
                      close();
                    }}
                  >
                    Unpublish
                  </MenuItem>
                )}

                <ConfirmDelete
                  key="delete"
                  onSubmit={ev => {
                    onClickDelete(ev);
                    close();
                  }}
                >
                  {confirm => (
                    <MenuItem
                      onClick={() => {
                        confirm.open();
                      }}
                    >
                      Delete
                    </MenuItem>
                  )}
                </ConfirmDelete>
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
