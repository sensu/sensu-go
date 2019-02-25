import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import ConfirmDelete from "/app/component/partial/ConfirmDelete";
import DeleteMenuItem from "/app/component/partial/ToolbarMenuItems/Delete";
import ListHeader from "/app/component/partial/ListHeader";
import ListSortSelector from "/app/component/partial/ListSortSelector";
import Select, { Option } from "/app/component/partial/ToolbarMenuItems/Select";
import PublishMenuItem from "/app/component/partial/ToolbarMenuItems/Publish";
import SilenceMenuItem from "/app/component/partial/ToolbarMenuItems/Silence";
import ToolbarMenu from "/app/component/partial/ToolbarMenu";
import UnpublishMenuItem from "/app/component/partial/ToolbarMenuItems/Unpublish";
import UnsilenceMenuItem from "/app/component/partial/ToolbarMenuItems/Unsilence";
import QueueMenuItem from "/app/component/partial/ToolbarMenuItems/QueueExecution";

class ChecksListHeader extends React.PureComponent {
  static propTypes = {
    editable: PropTypes.bool.isRequired,
    namespace: PropTypes.object,
    onChangeQuery: PropTypes.func.isRequired,
    onClickClearSilences: PropTypes.func.isRequired,
    onClickDelete: PropTypes.func.isRequired,
    onClickExecute: PropTypes.func.isRequired,
    onClickSetPublish: PropTypes.func.isRequired,
    onClickSilence: PropTypes.func.isRequired,
    order: PropTypes.string.isRequired,
    rowCount: PropTypes.number.isRequired,
    selectedItems: PropTypes.array.isRequired,
    toggleSelectedItems: PropTypes.func.isRequired,
  };

  static defaultProps = {
    namespace: null,
  };

  static fragments = {
    namespace: gql`
      fragment ChecksListHeader_namespace on Namespace {
        subscriptions(orderBy: OCCURRENCES) {
          values(limit: 25)
        }
      }
    `,
    check: gql`
      fragment ChecksListHeader_check on CheckConfig {
        id
        publish
        silences {
          id
        }
      }
    `,
  };

  updateFilter = val => {
    this.props.onChangeQuery({
      filter: `subscriptions.indexOf("${val}") >= 0`,
    });
  };

  renderActions = () => {
    const { namespace, onChangeQuery, order } = this.props;
    const subscriptions = namespace ? namespace.subscriptions.values : [];

    return (
      <ToolbarMenu>
        <ToolbarMenu.Item id="filter-by-subscription" visible="if-room">
          <Select title="Subscription" onChange={this.updateFilter}>
            {subscriptions.map(val => (
              <Option key={val} value={val} />
            ))}
          </Select>
        </ToolbarMenu.Item>
        <ToolbarMenu.Item id="sort" visible="if-room">
          <ListSortSelector
            options={[{ label: "Name", value: "NAME" }]}
            onChangeQuery={onChangeQuery}
            value={order}
          />
        </ToolbarMenu.Item>
      </ToolbarMenu>
    );
  };

  renderBulkActions = () => {
    const { selectedItems } = this.props;

    const selectedCount = selectedItems.length;
    const selectedSilenced = selectedItems.filter(en => en.silences.length > 0);
    const selectedPublished = selectedItems.filter(ch => ch.publish === true);
    const selectedNonKeepalives = selectedItems.filter(
      ch => ch.name !== "keepalive",
    );

    const allSelectedSilenced = selectedSilenced.length === selectedCount;
    const allSelectedUnsilenced = selectedSilenced.length === 0;
    const published = selectedCount === selectedPublished.length;

    return (
      <ToolbarMenu>
        <ToolbarMenu.Item id="queue" visible="always">
          <QueueMenuItem
            disabled={selectedNonKeepalives.length === 0}
            onClick={this.props.onClickExecute}
            description="Queue an adhoc execution of the selected checks."
          />
        </ToolbarMenu.Item>
        <ToolbarMenu.Item
          id="silence"
          visible={allSelectedSilenced ? "never" : "if-room"}
        >
          <SilenceMenuItem
            disabled={allSelectedSilenced}
            onClick={this.props.onClickSilence}
          />
        </ToolbarMenu.Item>
        <ToolbarMenu.Item
          id="unsilence"
          visible={allSelectedUnsilenced ? "never" : "if-room"}
        >
          <UnsilenceMenuItem
            disabled={allSelectedUnsilenced}
            onClick={this.props.onClickClearSilences}
          />
        </ToolbarMenu.Item>
        {!published ? (
          <ToolbarMenu.Item id="publish" visible={"if-room"}>
            <PublishMenuItem
              description="Publish selected checks."
              onClick={() => this.props.onClickSetPublish(true)}
            />
          </ToolbarMenu.Item>
        ) : (
          <ToolbarMenu.Item id="unpublish" visible={"if-room"}>
            <UnpublishMenuItem
              description="Unpublish selected checks."
              onClick={() => this.props.onClickSetPublish(false)}
            />
          </ToolbarMenu.Item>
        )}
        <ToolbarMenu.Item id="delete" visible="never">
          {menu => (
            <ConfirmDelete
              identifier={
                selectedCount > 1 ? `${selectedCount} checks` : "this check"
              }
              onSubmit={ev => {
                this.props.onClickDelete(ev);
                menu.close();
              }}
            >
              {dialog => <DeleteMenuItem onClick={dialog.open} />}
            </ConfirmDelete>
          )}
        </ToolbarMenu.Item>
      </ToolbarMenu>
    );
  };

  render() {
    const {
      editable,
      selectedItems,
      toggleSelectedItems,
      rowCount,
    } = this.props;

    return (
      <ListHeader
        sticky
        editable={editable}
        selectedCount={selectedItems.length}
        rowCount={rowCount}
        onClickSelect={toggleSelectedItems}
        renderActions={this.renderActions}
        renderBulkActions={this.renderBulkActions}
      />
    );
  }
}

export default ChecksListHeader;
