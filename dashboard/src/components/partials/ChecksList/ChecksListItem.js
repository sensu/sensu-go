import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Checkbox from "@material-ui/core/Checkbox";
import Code from "/components/Code";
import CodeHighlight from "/components/CodeHighlight/CodeHighlight";
import ConfirmDelete from "/components/partials/ConfirmDelete";
import DeleteMenuItem from "/components/partials/ToolbarMenuItems/Delete";
import HoverController from "/components/controller/HoverController";
import NamespaceLink from "/components/util/NamespaceLink";
import PublishMenuItem from "/components/partials/ToolbarMenuItems/Publish";
import ResourceDetails from "/components/partials/ResourceDetails";
import SilenceIcon from "/icons/Silence";
import SilenceMenuItem from "/components/partials/ToolbarMenuItems/Silence";
import TableCell from "@material-ui/core/TableCell";
import TableOverflowCell from "/components/partials/TableOverflowCell";
import TableSelectableRow from "/components/partials/TableSelectableRow";
import { FloatingTableToolbarCell } from "/components/partials/TableToolbarCell";
import ToolbarMenu from "/components/partials/ToolbarMenu";
import UnsilenceMenuItem from "/components/partials/ToolbarMenuItems/Unsilence";
import UnpublishMenuItem from "/components/partials/ToolbarMenuItems/Unpublish";
import QueueMenuItem from "/components/partials/ToolbarMenuItems/QueueExecution";

import CheckSchedule from "./CheckSchedule";

class CheckListItem extends React.Component {
  static propTypes = {
    editable: PropTypes.bool.isRequired,
    editing: PropTypes.bool.isRequired,
    check: PropTypes.object.isRequired,
    hovered: PropTypes.bool.isRequired,
    onHover: PropTypes.func.isRequired,
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
        namespace
        command
        isSilenced
        publish
        ...CheckSchedule_check
      }

      ${CheckSchedule.fragments.check}
    `,
  };

  render() {
    const { editable, editing, check, selected, onChangeSelected } = this.props;

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
                  <CodeHighlight
                    language="bash"
                    code={check.command}
                    component={Code}
                  />
                  <br />
                  <CheckSchedule check={check} />
                </React.Fragment>
              }
            />
          </TableOverflowCell>

          <FloatingTableToolbarCell
            hovered={this.props.hovered}
            disabled={!editable || editing}
          >
            {() => (
              <ToolbarMenu>
                <ToolbarMenu.Item id="queue" visible="never">
                  <QueueMenuItem
                    disabled={check.name === "keepalive"}
                    onClick={this.props.onClickExecute}
                  />
                </ToolbarMenu.Item>
                <ToolbarMenu.Item id="silence" visible="never">
                  <SilenceMenuItem
                    disabled={!!check.isSilenced}
                    onClick={this.props.onClickSilence}
                  />
                </ToolbarMenu.Item>
                <ToolbarMenu.Item id="unsilence" visible="never">
                  <UnsilenceMenuItem
                    disabled={!check.isSilenced}
                    onClick={this.props.onClickClearSilences}
                  />
                </ToolbarMenu.Item>
                {!check.publish ? (
                  <ToolbarMenu.Item id="publish" visible="never">
                    <PublishMenuItem
                      description="Publish check"
                      onClick={() => this.props.onClickSetPublish(true)}
                    />
                  </ToolbarMenu.Item>
                ) : (
                  <ToolbarMenu.Item id="unpublish" visible="never">
                    <UnpublishMenuItem
                      delete
                      description="Unpublish check"
                      onClick={() => this.props.onClickSetPublish(false)}
                    />
                  </ToolbarMenu.Item>
                )}
                <ToolbarMenu.Item id="delete" visible="never">
                  <ConfirmDelete onSubmit={this.props.onClickDelete}>
                    {dialog => <DeleteMenuItem onClick={dialog.open} />}
                  </ConfirmDelete>
                </ToolbarMenu.Item>
              </ToolbarMenu>
            )}
          </FloatingTableToolbarCell>
        </TableSelectableRow>
      </HoverController>
    );
  }
}

export default CheckListItem;
