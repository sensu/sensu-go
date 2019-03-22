import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import ClearSilenceAction from "/app/component/partial/ClearSilenceAction";
import DeleteMenuItem from "/app/component/partial/ToolbarMenuItems/Delete";
import PublishMenuItem from "/app/component/partial/ToolbarMenuItems/Publish";
import SilenceMenuItem from "/app/component/partial/ToolbarMenuItems/Silence";
import Toolbar from "/app/component/partial/Toolbar";
import ToolbarMenu from "/app/component/partial/ToolbarMenu";
import UnpublishMenuItem from "/app/component/partial/ToolbarMenuItems/Unpublish";
import UnsilenceMenuItem from "/app/component/partial/ToolbarMenuItems/Unsilence";
import QueueMenuItem from "/app/component/partial/ToolbarMenuItems/QueueExecution";

import DeleteAction from "./CheckDetailsDeleteAction";
import ExecuteAction from "./CheckDetailsExecuteAction";
import PublishAction from "./CheckDetailsPublishAction";
import UnpublishAction from "./CheckDetailsUnpublishAction";
import SilenceAction from "./CheckDetailsSilenceAction";

class CheckDetailsToolbar extends React.Component {
  static propTypes = {
    check: PropTypes.object,
    refetch: PropTypes.func,
  };

  static defaultProps = {
    check: null,
    refetch: () => null,
  };

  static fragments = {
    check: gql`
      fragment CheckDetailsToolbar_check on CheckConfig {
        isSilenced

        ...CheckDetailsDeleteAction_check
        ...CheckDetailsExecuteAction_check
        ...CheckDetailsSilenceAction_check
        ...ClearSilenceAction_record
      }

      ${DeleteAction.fragments.check}
      ${ExecuteAction.fragments.check}
      ${SilenceAction.fragments.check}
      ${ClearSilenceAction.fragments.record}
    `,
  };

  render() {
    const { check, refetch } = this.props;

    return (
      <Toolbar
        right={
          <ToolbarMenu>
            <ToolbarMenu.Item id="execute " visible="always">
              <ExecuteAction check={check}>
                {handler => <QueueMenuItem onClick={handler} />}
              </ExecuteAction>
            </ToolbarMenu.Item>
            <ToolbarMenu.Item
              id="silence"
              visible={check.isSilenced ? "never" : "if-room"}
            >
              <SilenceAction check={check} onDone={refetch}>
                {dialog => (
                  <SilenceMenuItem
                    onClick={dialog.open}
                    disabled={dialog.canOpen}
                  />
                )}
              </SilenceAction>
            </ToolbarMenu.Item>
            <ToolbarMenu.Item
              id="unsilence"
              visible={check.isSilenced ? "if-room" : "never"}
            >
              <ClearSilenceAction record={check} onDone={refetch}>
                {dialog => (
                  <UnsilenceMenuItem
                    onClick={dialog.open}
                    disabled={!dialog.canOpen}
                  />
                )}
              </ClearSilenceAction>
            </ToolbarMenu.Item>
            <ToolbarMenu.Item
              id={check.publish ? "unpublish" : "publish"}
              visible="if-room"
            >
              {check.publish ? (
                <UnpublishAction check={check}>
                  {handler => <UnpublishMenuItem onClick={handler} />}
                </UnpublishAction>
              ) : (
                <PublishAction check={check}>
                  {handler => <PublishMenuItem onClick={handler} />}
                </PublishAction>
              )}
            </ToolbarMenu.Item>
            <ToolbarMenu.Item id="delete" visible="never">
              <DeleteAction check={check}>
                {handler => <DeleteMenuItem onClick={handler} />}
              </DeleteAction>
            </ToolbarMenu.Item>
          </ToolbarMenu>
        }
      />
    );
  }
}

export default CheckDetailsToolbar;
