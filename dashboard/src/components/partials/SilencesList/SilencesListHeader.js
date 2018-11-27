import React from "react";
import PropTypes from "prop-types";

import UnsilenceMenuItem from "/components/partials/ToolbarMenuItems/Unsilence";
import ListHeader from "/components/partials/ListHeader";
import ListSortSelector from "/components/partials/ListSortSelector";
import ToolbarMenu from "/components/partials/ToolbarMenu";

class SilencesListHeader extends React.PureComponent {
  static propTypes = {
    onClickClearSilences: PropTypes.func.isRequired,
    onClickSelect: PropTypes.func.isRequired,
    onChangeQuery: PropTypes.func.isRequired,
    order: PropTypes.string.isRequired,
    selectedItems: PropTypes.array,
    rowCount: PropTypes.number,
  };

  static defaultProps = {
    rowCount: 0,
    selectedItems: [],
  };

  state = { openDialog: false };

  renderActions = () => {
    const { onChangeQuery, order } = this.props;

    return (
      <ToolbarMenu>
        <ToolbarMenu.Item id="sort" visible="always">
          <ListSortSelector
            onChangeQuery={onChangeQuery}
            options={[
              { label: "Name", value: "ID" },
              { label: "Start Date", value: "BEGIN" },
            ]}
            value={order}
          />
        </ToolbarMenu.Item>
      </ToolbarMenu>
    );
  };

  renderBulkActions = () => (
    <ToolbarMenu>
      <ToolbarMenu.Item id="clearSilence" visible="always">
        <UnsilenceMenuItem onClick={this.props.onClickClearSilences} />
      </ToolbarMenu.Item>
    </ToolbarMenu>
  );

  render() {
    const { onClickSelect, selectedItems, rowCount } = this.props;

    return (
      <ListHeader
        sticky
        selectedCount={selectedItems.length}
        onClickSelect={onClickSelect}
        renderActions={this.renderActions}
        renderBulkActions={this.renderBulkActions}
        rowCount={rowCount}
      />
    );
  }
}

export default SilencesListHeader;
