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

import ConfirmDelete from "/components/partials/ConfirmDelete";
import ResourceDetails from "/components/partials/ResourceDetails";
import TableOverflowCell from "/components/partials/TableOverflowCell";
import TableSelectableRow from "/components/partials/TableSelectableRow";

import EntityStatusDescriptor from "/components/partials/EntityStatusDescriptor";
import CheckStatusIcon from "/components/CheckStatusIcon";
import NamespaceLink from "/components/util/NamespaceLink";

class EntitiesListItem extends React.PureComponent {
  static propTypes = {
    entity: PropTypes.object.isRequired,
    selected: PropTypes.bool,
    onChangeSelected: PropTypes.func,
    onClickClearSilence: PropTypes.func,
    onClickDelete: PropTypes.func,
    onClickSilence: PropTypes.func,
  };

  static defaultProps = {
    selected: undefined,
    onChangeSelected: ev => ev,
    onClickClearSilence: ev => ev,
    onClickDelete: ev => ev,
    onClickSilence: ev => ev,
  };

  static fragments = {
    entity: gql`
      fragment EntitiesListItem_entity on Entity {
        id
        name
        status
        isSilenced
        system {
          platform
          platformVersion
        }
        ...EntityStatusDescriptor_entity
      }

      ${EntityStatusDescriptor.fragments.entity}
    `,
  };

  renderMenu = ({ close, anchorEl }) => {
    const {
      entity,
      onClickClearSilence,
      onClickDelete,
      onClickSilence,
    } = this.props;

    return (
      <Menu open onClose={close} anchorEl={anchorEl}>
        {entity.isSilenced ? (
          <MenuItem
            onClick={ev => {
              onClickClearSilence(ev);
              close(ev);
            }}
          >
            Unsilence
          </MenuItem>
        ) : (
          <MenuItem
            onClick={ev => {
              onClickSilence(ev);
              close(ev);
            }}
          >
            Silence
          </MenuItem>
        )}
        <ConfirmDelete key="delete" onSubmit={onClickDelete}>
          {confirm => (
            <MenuItem
              onClick={() => {
                confirm.open();
                close();
              }}
            >
              Delete
            </MenuItem>
          )}
        </ConfirmDelete>
      </Menu>
    );
  };

  render() {
    const { entity, selected, onChangeSelected } = this.props;

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
            icon={
              <CheckStatusIcon
                statusCode={entity.status}
                silenced={entity.isSilenced}
              />
            }
            title={
              <NamespaceLink
                namespace={entity.namespace}
                to={`/entities/${entity.name}`}
              >
                <strong>{entity.name}</strong> {entity.system.platform}{" "}
                {entity.system.platformVersion}
              </NamespaceLink>
            }
            details={<EntityStatusDescriptor entity={entity} />}
          />
        </TableOverflowCell>
        <TableCell padding="checkbox">
          <MenuController renderMenu={this.renderMenu}>
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

export default EntitiesListItem;
