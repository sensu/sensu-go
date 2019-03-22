import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Checkbox from "@material-ui/core/Checkbox";
import TableCell from "@material-ui/core/TableCell";

import SilenceIcon from "/lib/component/icon/Silence";
import CheckStatusIcon from "/lib/component/base/CheckStatusIcon";
import HoverController from "/lib/component/controller/HoverController";
import NamespaceLink from "/lib/component/util/NamespaceLink";

import ConfirmDelete from "/app/component/partial/ConfirmDelete";
import DeleteMenuItem from "/app/component/partial/ToolbarMenuItems/Delete";
import QueueMenuItem from "/app/component/partial/ToolbarMenuItems/QueueExecution";
import ResolveMenuItem from "/app/component/partial/ToolbarMenuItems/Resolve";
import Select, { Option } from "/app/component/partial/ToolbarMenuItems/Select";
import UnsilenceMenuItem from "/app/component/partial/ToolbarMenuItems/Unsilence";
import ToolbarMenu from "/app/component/partial/ToolbarMenu";

import ResourceDetails from "/app/component/partial/ResourceDetails";
import TableOverflowCell from "/app/component/partial/TableOverflowCell";
import { FloatingTableToolbarCell } from "/app/component/partial/TableToolbarCell";
import TableSelectableRow from "/app/component/partial/TableSelectableRow";

import EventStatusDescriptor from "/app/component/partial/EventStatusDescriptor";

class EventListItem extends React.Component {
  static propTypes = {
    editable: PropTypes.bool,
    editing: PropTypes.bool,
    hovered: PropTypes.bool.isRequired,
    selected: PropTypes.bool.isRequired,
    onHover: PropTypes.func.isRequired,
    onChangeSelected: PropTypes.func.isRequired,
    onClickClearSilences: PropTypes.func.isRequired,
    onClickDelete: PropTypes.func.isRequired,
    onClickSilencePair: PropTypes.func.isRequired,
    onClickSilenceEntity: PropTypes.func.isRequired,
    onClickSilenceCheck: PropTypes.func.isRequired,
    onClickResolve: PropTypes.func.isRequired,
    onClickRerun: PropTypes.func.isRequired,
    event: PropTypes.object.isRequired,
  };

  static defaultProps = {
    editable: true,
    editing: false,
  };

  static fragments = {
    event: gql`
      fragment EventsListItem_event on Event {
        isSilenced
        isNewIncident
        timestamp
        check {
          name
          status
          ...EventStatusDescriptor_check
        }
        entity {
          name
        }
        namespace
        ...EventStatusDescriptor_event
      }

      ${EventStatusDescriptor.fragments.check}
      ${EventStatusDescriptor.fragments.event}
    `,
  };

  handleClickCheckbox = () => {
    this.props.onChangeSelected(!this.props.selected);
  };

  render() {
    const { editable, editing, selected, event } = this.props;
    const { entity, check, timestamp } = event;

    // Try to determine if the failing check just started failing and if so
    // highlight the row.
    const isNewIncident =
      event.isNewIncident &&
      new Date(new Date(timestamp).valueOf() + 2500) >= new Date();

    return (
      <HoverController onHover={this.props.onHover}>
        <TableSelectableRow selected={selected} highlight={isNewIncident}>
          {editable && (
            <TableCell padding="checkbox">
              <Checkbox
                color="primary"
                checked={selected}
                onChange={this.handleClickCheckbox}
              />
            </TableCell>
          )}

          <TableOverflowCell>
            <ResourceDetails
              icon={
                event.check && (
                  <CheckStatusIcon
                    statusCode={event.check.status}
                    silenced={event.isSilenced}
                  />
                )
              }
              title={
                <NamespaceLink
                  namespace={event.namespace}
                  to={`/events/${entity.name}/${check.name}`}
                >
                  <strong>
                    {entity.name} › {check.name}
                  </strong>
                </NamespaceLink>
              }
              details={
                <EventStatusDescriptor event={event} check={event.check} />
              }
            />
          </TableOverflowCell>

          <FloatingTableToolbarCell
            hovered={this.props.hovered}
            disabled={!editable || editing}
          >
            {() => (
              <ToolbarMenu>
                <ToolbarMenu.Item id="resolve" visible="always">
                  <ResolveMenuItem
                    iconOnly
                    disabled={event.status === 0}
                    onClick={this.props.onClickResolve}
                  />
                </ToolbarMenu.Item>
                <ToolbarMenu.Item id="re-run" visible="never">
                  <QueueMenuItem
                    disabled={event.check.name === "keepalive"}
                    onClick={this.props.onClickRerun}
                  />
                </ToolbarMenu.Item>
                <ToolbarMenu.Item id="silence" visible="never">
                  <Select
                    disabled={event.isSilenced}
                    icon={<SilenceIcon />}
                    primary="Silence"
                    onChange={sl => {
                      if (sl === "check") {
                        this.props.onClickSilenceCheck();
                      } else if (sl === "entity") {
                        this.props.onClickSilenceEntity();
                      } else {
                        this.props.onClickSilencePair();
                      }
                    }}
                  >
                    <Option value="check">Check</Option>
                    <Option value="entity">Entity</Option>
                    <Option value="both">Both</Option>
                  </Select>
                </ToolbarMenu.Item>
                <ToolbarMenu.Item id="unsilenced" visible="never">
                  <UnsilenceMenuItem
                    disabled={!event.isSilenced}
                    onClick={this.props.onClickClearSilences}
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

export default EventListItem;
