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

class EntitiesListHeader extends React.PureComponent {
  static propTypes = {
    environment: PropTypes.object,
    onChangeQuery: PropTypes.func.isRequired,
    onClickClearSilences: PropTypes.func.isRequired,
    onClickDelete: PropTypes.func.isRequired,
    onClickSelect: PropTypes.func.isRequired,
    onClickSilence: PropTypes.func.isRequired,
    rowCount: PropTypes.number.isRequired,
    selectedItems: PropTypes.array.isRequired,
    order: PropTypes.string.isRequired,
  };

  static defaultProps = {
    onClickSelect: () => {},
    onChangeFilter: () => {},
    onChangeSort: () => {},
    onSubmitDelete: () => {},
    environment: undefined,
  };

  static fragments = {
    environment: gql`
      fragment EntitiesListHeader_environment on Environment {
        subscriptions(orderBy: OCCURRENCES, omitEntity: true) {
          values(limit: 25)
        }
      }
    `,
  };

  updateFilter = val => {
    const filter = `'${val}' IN Subscriptions`;
    this.props.onChangeQuery({ filter });
  };

  renderActions = () => {
    const { environment: env, onChangeQuery, order } = this.props;
    const subs = env ? env.subscriptions.values : [];

    return (
      <ToolbarMenu>
        <ToolbarMenu.Item id="filter-by-subscriptions" visible="if-room">
          <Select title="Subscription" onChange={this.updateFilter}>
            {subs.map(v => <Option key={v} value={v} />)}
          </Select>
        </ToolbarMenu.Item>
        <ToolbarMenu.Item id="sort" visible="always">
          <ListSortSelector
            options={[{ label: "Name", value: "ID" }]}
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
        <ToolbarMenu.Item
          id="silence"
          visible={allSelectedSilenced ? "never" : "always"}
        >
          <SilenceMenuItem
            description="Create a silence targeting selected entities."
            disabled={allSelectedSilenced}
            onClick={this.props.onClickSilence}
          />
        </ToolbarMenu.Item>

        <ToolbarMenu.Item
          id="unsilence"
          visible={allSelectedUnsilenced ? "never" : "if-room"}
        >
          <UnsilenceMenuItem
            description="Clear silences associated with selected entities."
            disabled={allSelectedUnsilenced}
            onClick={this.props.onClickClearSilences}
          />
        </ToolbarMenu.Item>

        <ToolbarMenu.Item id="delete" visible="never">
          {menu => (
            <ConfirmDelete
              identifier={`${selectedCount} ${
                selectedCount === 1 ? "entity" : "entities"
              }`}
              onSubmit={() => {
                this.props.onClickDelete();
                menu.close();
              }}
            >
              {confirm => (
                <DeleteMenuItem
                  autoClose={false}
                  title="Delete…"
                  onClick={confirm.open}
                />
              )}
            </ConfirmDelete>
          )}
        </ToolbarMenu.Item>
      </ToolbarMenu>
    );
  };

  render() {
    const { onClickSelect, selectedItems, rowCount } = this.props;
    const selectedCount = selectedItems.length;

    return (
      <ListHeader
        sticky
        selectedCount={selectedCount}
        rowCount={rowCount}
        onClickSelect={onClickSelect}
        renderBulkActions={this.renderBulkActions}
        renderActions={this.renderActions}
      />
    );
  }
}

export default EntitiesListHeader;
