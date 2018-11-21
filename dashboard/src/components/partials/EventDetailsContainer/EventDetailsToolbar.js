import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import DeleteMenuItem from "/components/partials/ToolbarMenuItems/Delete";
import QueueMenuItem from "/components/partials/ToolbarMenuItems/QueueExecution";
import ResolveMenuItem from "/components/partials/ToolbarMenuItems/Resolve";
import Toolbar from "/components/partials/Toolbar";
import ToolbarMenu from "/components/partials/ToolbarMenu";

import DeleteAction from "./EventDetailsDeleteAction";
import ResolveAction from "./EventDetailsResolveAction";
import ReRunAction from "./EventDetailsReRunAction";

class EventDetailsToolbar extends React.Component {
  static propTypes = {
    event: PropTypes.object.isRequired,
  };

  static fragments = {
    event: gql`
      fragment EventDetailsToolbar_event on Event {
        ...EventDetailsDeleteAction_event
        ...EventDetailsResolveAction_event
        ...EventDetailsReRunAction_event
      }

      ${DeleteAction.fragments.event}
      ${ResolveAction.fragments.event}
      ${ReRunAction.fragments.event}
    `,
  };

  render() {
    const { event } = this.props;

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
