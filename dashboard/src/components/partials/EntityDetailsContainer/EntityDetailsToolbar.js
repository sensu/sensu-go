import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import DeleteMenuItem from "/components/partials/ToolbarMenuItems/Delete";
import SilenceMenuItem from "/components/partials/ToolbarMenuItems/Silence";
import Toolbar from "/components/partials/Toolbar";
import ToolbarMenu from "/components/partials/ToolbarMenu";
import UnsilenceMenuItem from "/components/partials/ToolbarMenuItems/Unsilence";

import DeleteAction from "./EntityDetailsDeleteAction";
import SilenceAction from "./EntityDetailsSilenceAction";
import ClearSilenceAction from "./EntityDetailsClearSilenceAction";

class EntityDetailsToolbar extends React.Component {
  static propTypes = {
    entity: PropTypes.object.isRequired,
    refetch: PropTypes.func.isRequired,
  };

  static fragments = {
    entity: gql`
      fragment EntityDetailsToolbar_entity on Entity {
        isSilenced

        ...EntityDetailsDeleteAction_entity
        ...EntityDetailsSilenceAction_entity
        ...EntityDetailsClearSilenceAction_entity
      }

      ${DeleteAction.fragments.entity}
      ${SilenceAction.fragments.entity}
      ${ClearSilenceAction.fragments.entity}
    `,
  };

  render() {
    const { entity, refetch } = this.props;

    return (
      <Toolbar
        right={
          <ToolbarMenu>
            <ToolbarMenu.Item
              id="silence"
              visible={entity.isSilenced ? "never" : "if-room"}
            >
              <SilenceAction entity={entity} onDone={refetch}>
                {dialog => (
                  <SilenceMenuItem
                    disabled={dialog.canOpen}
                    onClick={dialog.open}
                  />
                )}
              </SilenceAction>
            </ToolbarMenu.Item>

            <ToolbarMenu.Item
              id="unsilence"
              visible={entity.isSilenced ? "if-room" : "never"}
            >
              <ClearSilenceAction entity={entity} onDone={refetch}>
                {dialog => (
                  <UnsilenceMenuItem
                    disabled={dialog.canOpen}
                    onClick={dialog.open}
                  />
                )}
              </ClearSilenceAction>
            </ToolbarMenu.Item>

            <ToolbarMenu.Item id="delete" visible="never">
              <DeleteAction entity={entity}>
                {handler => <DeleteMenuItem onClick={handler} />}
              </DeleteAction>
            </ToolbarMenu.Item>
          </ToolbarMenu>
        }
      />
    );
  }
}

export default EntityDetailsToolbar;
