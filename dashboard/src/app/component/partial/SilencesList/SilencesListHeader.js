import React from "react";
import PropTypes from "prop-types";

import UnsilenceMenuItem from "/app/component/partial/ToolbarMenuItems/Unsilence";
import ListHeader from "/app/component/partial/ListHeader";
import ListSortSelector from "/app/component/partial/ListSortSelector";
import ToolbarMenu from "/app/component/partial/ToolbarMenu";

class SilencesListHeader extends React.PureComponent {
  static propTypes = {
    editable: PropTypes.bool.isRequired,
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
    const { editable, onClickSelect, selectedItems, rowCount } = this.props;

    return (
      <ListHeader
        sticky
        editable={editable}
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
