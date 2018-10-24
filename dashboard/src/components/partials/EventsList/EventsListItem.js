import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Checkbox from "@material-ui/core/Checkbox";
import TableCell from "@material-ui/core/TableCell";

import ConfirmDelete from "/components/partials/ConfirmDelete";
import DeleteMenuItem from "/components/partials/ToolbarMenuItems/Delete";
import QueueMenuItem from "/components/partials/ToolbarMenuItems/QueueExecution";
import ResolveMenuItem from "/components/partials/ToolbarMenuItems/Resolve";
import Select, { Option } from "/components/partials/ToolbarMenuItems/Select";
import SilenceIcon from "/icons/Silence";
import UnsilenceMenuItem from "/components/partials/ToolbarMenuItems/Unsilence";
import ToolbarMenu from "/components/partials/ToolbarMenu";

import ResourceDetails from "/components/partials/ResourceDetails";
import TableOverflowCell from "/components/partials/TableOverflowCell";
import TableSelectableRow from "/components/partials/TableSelectableRow";

import EventStatusDescriptor from "/components/partials/EventStatusDescriptor";
import NamespaceLink from "/components/util/NamespaceLink";
import CheckStatusIcon from "/components/CheckStatusIcon";

class EventListItem extends React.Component {
  static propTypes = {
    selected: PropTypes.bool.isRequired,
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
    const { selected, event } = this.props;
    const { entity, check, timestamp } = event;

    // Try to determine if the failing check just started failing and if so
    // highlight the row.
    const isNewIncident =
      event.isNewIncident &&
      new Date(new Date(timestamp).valueOf() + 2500) >= new Date();

    return (
      <TableSelectableRow selected={selected} highlight={isNewIncident}>
        <TableCell padding="checkbox">
          <Checkbox
            color="primary"
            checked={selected}
            onChange={this.handleClickCheckbox}
          />
        </TableCell>

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

        <TableCell padding="checkbox">
          <ToolbarMenu>
            <ToolbarMenu.Item id="resolve" visible="never">
              <ResolveMenuItem onClick={this.props.onClickResolve} />
            </ToolbarMenu.Item>
            <ToolbarMenu.Item id="re-run" visible="never">
              <QueueMenuItem
                title="Re-run Check"
                onClick={this.props.onClickRerun}
              />
            </ToolbarMenu.Item>
            <ToolbarMenu.Item id="re-run" visible="never">
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
            <ToolbarMenu.Item id="re-run" visible="never">
              <UnsilenceMenuItem
                disabled={!event.isSilenced}
                onClick={this.props.onClickClearSilences}
              />
            </ToolbarMenu.Item>
            <ToolbarMenu.Item id="re-run" visible="never">
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
        </TableCell>
      </TableSelectableRow>
    );
  }
}

export default EventListItem;
