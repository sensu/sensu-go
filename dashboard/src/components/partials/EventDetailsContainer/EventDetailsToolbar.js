import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import DeleteMenuItem from "/components/partials/ToolbarMenuItems/Delete";
import QueueMenuItem from "/components/partials/ToolbarMenuItems/QueueExecution";
import ResolveMenuItem from "/components/partials/ToolbarMenuItems/Resolve";
import SilenceMenuItem from "/components/partials/ToolbarMenuItems/Silence";
import UnsilenceMenuItem from "/components/partials/ToolbarMenuItems/Unsilence";
import Toolbar from "/components/partials/Toolbar";
import ToolbarMenu from "/components/partials/ToolbarMenu";

import DeleteAction from "./EventDetailsDeleteAction";
import ResolveAction from "./EventDetailsResolveAction";
import ReRunAction from "./EventDetailsReRunAction";
import SilenceAction from "./EventDetailsSilenceAction";
import ClearSilenceAction from "./EventDetailsClearSilenceAction";

class EventDetailsToolbar extends React.Component {
  static propTypes = {
    event: PropTypes.object.isRequired,
    refetch: PropTypes.func.isRequired,
  };

  static defaultProps = {
    refetch: () => null,
  };

  static fragments = {
    event: gql`
      fragment EventDetailsToolbar_event on Event {
        ...EventDetailsDeleteAction_event
        ...EventDetailsResolveAction_event
        ...EventDetailsReRunAction_event
        ...EventDetailsSilenceAction_event
        ...EventDetailsClearSilenceAction_event

        check {
          silenced
        }
      }

      ${DeleteAction.fragments.event}
      ${ResolveAction.fragments.event}
      ${ReRunAction.fragments.event}
      ${SilenceAction.fragments.event}
      ${ClearSilenceAction.fragments.event}
    `,
  };

  render() {
    const { event, refetch } = this.props;

    return (
      <Toolbar
        right={
          <ToolbarMenu fillWidth>
            <ToolbarMenu.Item id="resolve" visible="always">
              <ResolveAction event={event}>
                {({ resolve, canResolve }) => (
                  <ResolveMenuItem disabled={!canResolve} onClick={resolve} />
                )}
              </ResolveAction>
            </ToolbarMenu.Item>
            <ToolbarMenu.Item id="re-run" visible="if-room">
              {event.check.name !== "keepalive" && (
                <ReRunAction event={event}>
                  {run => (
                    <QueueMenuItem
                      title="Re-run Check"
                      titleCondensed="Re-run"
                      onClick={run}
                    />
                  )}
                </ReRunAction>
              )}
            </ToolbarMenu.Item>
            <ToolbarMenu.Item id="silence" visible="if-room">
              {event.check.silenced.length === 0 ? (
                <SilenceAction event={event}>
                  {({ open }) => <SilenceMenuItem onClick={open} />}
                </SilenceAction>
              ) : (
                <ClearSilenceAction event={event} onDone={refetch}>
                  {({ open }) => <UnsilenceMenuItem onClick={open} />}
                </ClearSilenceAction>
              )}
            </ToolbarMenu.Item>
            <ToolbarMenu.Item id="delete" visible="never">
              <DeleteAction event={event}>
                {handler => <DeleteMenuItem onClick={handler} />}
              </DeleteAction>
            </ToolbarMenu.Item>
          </ToolbarMenu>
        }
      />
    );
  }
}

export default EventDetailsToolbar;
