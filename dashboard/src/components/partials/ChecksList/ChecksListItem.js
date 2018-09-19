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
import CheckPublishInfo from "/components/CheckPublishInfo";

class CheckListItem extends React.Component {
  static propTypes = {
    check: PropTypes.object.isRequired,
    selected: PropTypes.bool.isRequired,
    onChangeSelected: PropTypes.func.isRequired,
    onClickClearSilences: PropTypes.func.isRequired,
    onClickDelete: PropTypes.func.isRequired,
    onClickExecute: PropTypes.func.isRequired,
    onClickSilence: PropTypes.func.isRequired,
    onClickPublish: PropTypes.func.isRequired,
    onClickUnpublish: PropTypes.func.isRequired,
  };

  static fragments = {
    check: gql`
      fragment ChecksListItem_check on CheckConfig {
        name
        command
        subscriptions
        interval
        cron
        isSilenced
        publish
        namespace {
          organization
          environment
        }
      }
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
      onClickSilence,
      onClickPublish,
      onClickUnpublish,
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
            details={<CheckPublishInfo check={check} />}
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
                      onClickPublish();
                      close();
                    }}
                  >
                    Publish
                  </MenuItem>
                )}
                {check.publish && (
                  <MenuItem
                    onClick={() => {
                      onClickUnpublish();
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
