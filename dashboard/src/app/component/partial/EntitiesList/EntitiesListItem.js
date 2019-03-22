import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Checkbox from "@material-ui/core/Checkbox";
import TableCell from "@material-ui/core/TableCell";

import ConfirmDelete from "/app/component/partial/ConfirmDelete";
import DeleteMenuItem from "/app/component/partial/ToolbarMenuItems/Delete";
import SilenceMenuItem from "/app/component/partial/ToolbarMenuItems/Silence";
import UnsilenceMenuItem from "/app/component/partial/ToolbarMenuItems/Unsilence";
import ToolbarMenu from "/app/component/partial/ToolbarMenu";

import HoverController from "/lib/component/controller/HoverController";
import ResourceDetails from "/app/component/partial/ResourceDetails";
import TableOverflowCell from "/app/component/partial/TableOverflowCell";
import TableSelectableRow from "/app/component/partial/TableSelectableRow";
import { FloatingTableToolbarCell } from "/app/component/partial/TableToolbarCell";

import EntityStatusDescriptor from "/app/component/partial/EntityStatusDescriptor";
import CheckStatusIcon from "/lib/component/base/CheckStatusIcon";
import NamespaceLink from "/lib/component/util/NamespaceLink";

class EntitiesListItem extends React.PureComponent {
  static propTypes = {
    editable: PropTypes.bool.isRequired,
    editing: PropTypes.bool.isRequired,
    entity: PropTypes.object.isRequired,
    hovered: PropTypes.bool.isRequired,
    onHover: PropTypes.func.isRequired,
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

  render() {
    const {
      editable,
      editing,
      entity,
      selected,
      onChangeSelected,
    } = this.props;

    return (
      <HoverController onHover={this.props.onHover}>
        <TableSelectableRow selected={selected}>
          {editable && (
            <TableCell padding="checkbox">
              <Checkbox
                color="primary"
                checked={selected}
                onChange={e => onChangeSelected(e.target.checked)}
              />
            </TableCell>
          )}

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

          <FloatingTableToolbarCell
            hovered={this.props.hovered}
            disabled={!editable || editing}
          >
            {() => (
              <ToolbarMenu>
                <ToolbarMenu.Item id="silence" visible="never">
                  <SilenceMenuItem
                    disabled={entity.isSilenced}
                    onClick={this.props.onClickSilence}
                  />
                </ToolbarMenu.Item>
                <ToolbarMenu.Item id="unsilence" visible="never">
                  <UnsilenceMenuItem
                    disabled={!entity.isSilenced}
                    onClick={this.props.onClickClearSilence}
                  />
                </ToolbarMenu.Item>
                <ToolbarMenu.Item id="delete" visible="never">
                  {menu => (
                    <ConfirmDelete
                      onSubmit={() => {
                        this.props.onClickDelete();
                        menu.close();
                      }}
                    >
                      {dialog => (
                        <DeleteMenuItem
                          autoClose={false}
                          title="Delete…"
                          onClick={dialog.open}
                        />
                      )}
                    </ConfirmDelete>
                  )}
                </ToolbarMenu.Item>
              </ToolbarMenu>
            )}
          </FloatingTableToolbarCell>
        </TableSelectableRow>
      </HoverController>
    );
  }
}

export default EntitiesListItem;
