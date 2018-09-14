import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Checkbox from "@material-ui/core/Checkbox";
import Code from "/components/Code";
import ConfirmDelete from "/components/partials/ConfirmDelete";
import CronDescriptor from "/components/partials/CronDescriptor";
import DeleteMenuItem from "/components/partials/ToolbarMenuItems/Delete";
import NamespaceLink from "/components/util/NamespaceLink";
import ResourceDetails from "/components/partials/ResourceDetails";
import SilenceIcon from "/icons/Silence";
import SilenceMenuItem from "/components/partials/ToolbarMenuItems/Silence";
import TableCell from "@material-ui/core/TableCell";
import TableOverflowCell from "/components/partials/TableOverflowCell";
import TableSelectableRow from "/components/partials/TableSelectableRow";
import ToolbarMenu from "/components/partials/ToolbarMenu";
import UnsilenceMenuItem from "/components/partials/ToolbarMenuItems/Unsilence";
import QueueMenuItem from "/components/partials/ToolbarMenuItems/QueueExecution";

class CheckListItem extends React.Component {
  static propTypes = {
    check: PropTypes.object.isRequired,
    selected: PropTypes.bool.isRequired,
    onChangeSelected: PropTypes.func.isRequired,
    onClickClearSilences: PropTypes.func.isRequired,
    onClickDelete: PropTypes.func.isRequired,
    onClickExecute: PropTypes.func.isRequired,
    onClickSilence: PropTypes.func.isRequired,
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
        namespace {
          organization
          environment
        }
      }
    `,
  };

  render() {
    const { check, selected, onChangeSelected } = this.props;

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
            details={
              <React.Fragment>
                <Code>{check.command}</Code>
                <br />
                Executed{" "}
                <strong>
                  {check.interval ? (
                    `
                      every
                      ${check.interval}
                      ${check.interval === 1 ? "second" : "seconds"}
                    `
                  ) : (
                    <CronDescriptor expression={check.cron} />
                  )}
                </strong>{" "}
                by{" "}
                <strong>
                  {check.subscriptions.length}{" "}
                  {check.subscriptions.length === 1
                    ? "subscription"
                    : "subscriptions"}
                </strong>.
              </React.Fragment>
            }
          />
        </TableOverflowCell>

        <TableCell padding="checkbox">
          <ToolbarMenu>
            <ToolbarMenu.Item id="queue" visible="never">
              <QueueMenuItem onClick={this.props.onClickExecute} />
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
            <ToolbarMenu.Item id="delete" visible="never">
              <ConfirmDelete onSubmit={this.props.onClickDelete}>
                {dialog => <DeleteMenuItem onClick={dialog.open} />}
              </ConfirmDelete>
            </ToolbarMenu.Item>
          </ToolbarMenu>
        </TableCell>
      </TableSelectableRow>
    );
  }
}

export default CheckListItem;
