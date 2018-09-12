import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import ConfirmDelete from "/components/partials/ConfirmDelete";
import DeleteMenuItem from "/components/partials/ToolbarMenuItems/Delete";
import ListHeader from "/components/partials/ListHeader";
import ListSortSelector from "/components/partials/ListSortSelector";
import Select, { Option } from "/components/partials/ToolbarMenuItems/Select";
import SilenceMenuItem from "/components/partials/ToolbarMenuItems/Silence";
import ToolbarMenu from "/components/partials/ToolbarMenu";
import UnsilenceMenuItem from "/components/partials/ToolbarMenuItems/Unsilence";
import QueueMenuItem from "/components/partials/ToolbarMenuItems/QueueExecution";

class ChecksListHeader extends React.PureComponent {
  static propTypes = {
    environment: PropTypes.object,
    onChangeQuery: PropTypes.func.isRequired,
    onClickClearSilences: PropTypes.func.isRequired,
    onClickDelete: PropTypes.func.isRequired,
    onClickExecute: PropTypes.func.isRequired,
    onClickSilence: PropTypes.func.isRequired,
    order: PropTypes.string.isRequired,
    rowCount: PropTypes.number.isRequired,
    selectedItems: PropTypes.array.isRequired,
    toggleSelectedItems: PropTypes.func.isRequired,
  };

  static defaultProps = {
    environment: null,
  };

  static fragments = {
    environment: gql`
      fragment ChecksListHeader_environment on Environment {
        subscriptions(orderBy: OCCURRENCES) {
          values(limit: 25)
        }
      }
    `,
  };

  updateFilter = val => {
    this.props.onChangeQuery({ filter: `'${val}' IN Subscriptions` });
  };

  renderActions = () => {
    const { environment, onChangeQuery, order } = this.props;
    const subscriptions = environment ? environment.subscriptions.values : [];

    return (
      <ToolbarMenu>
        <ToolbarMenu.Item visible="if-room">
          <Select title="Subscription" onChange={this.updateFilter}>
            {subscriptions.map(val => <Option key={val} value={val} />)}
          </Select>
        </ToolbarMenu.Item>
        <ToolbarMenu.Item visible="if-room">
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
    const allSelectedSilenced = selectedSilenced.length === selectedCount;
    const allSelectedUnsilenced = selectedSilenced.length === 0;

    return (
      <ToolbarMenu>
        <ToolbarMenu.Item visible="always">
          <QueueMenuItem
            onClick={this.props.onClickExecute}
            description="Queue an adhoc execution of the selected checks."
          />
        </ToolbarMenu.Item>
        <ToolbarMenu.Item visible={allSelectedSilenced ? "never" : "if-room"}>
          <SilenceMenuItem
            disabled={allSelectedSilenced}
            onClick={this.props.onClickSilence}
          />
        </ToolbarMenu.Item>
        <ToolbarMenu.Item visible={allSelectedUnsilenced ? "never" : "if-room"}>
          <UnsilenceMenuItem
            disabled={allSelectedUnsilenced}
            onClick={this.props.onClickClearSilences}
          />
        </ToolbarMenu.Item>
        <ToolbarMenu.Item visible="never">
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
    const { selectedItems, toggleSelectedItems, rowCount } = this.props;
    const selectedCount = selectedItems.length;

    return (
      <ListHeader
        sticky
        selectedCount={selectedCount}
        rowCount={rowCount}
        onClickSelect={toggleSelectedItems}
        renderActions={this.renderActions}
        renderBulkActions={this.renderBulkActions}
      />
    );
  }
}

export default ChecksListHeader;
