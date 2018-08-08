import React from "react";
import PropTypes from "prop-types";

import CollapsingMenu from "/components/partials/CollapsingMenu";
import ConfirmDelete from "/components/partials/ConfirmDelete";
import DeleteIcon from "@material-ui/icons/Delete";
import ListHeader from "/components/partials/ListHeader";
import ListSortMenu from "/components/partials/ListSortMenu";
import SilenceIcon from "/icons/Silence";
import QueueIcon from "@material-ui/icons/Queue";
import UnsilenceIcon from "/icons/Unsilence";

class ChecksListHeader extends React.PureComponent {
  static propTypes = {
    onChangeQuery: PropTypes.func.isRequired,
    onClickClearSilences: PropTypes.func.isRequired,
    onClickDelete: PropTypes.func.isRequired,
    onClickExecute: PropTypes.func.isRequired,
    onClickSilence: PropTypes.func.isRequired,
    rowCount: PropTypes.number.isRequired,
    selectedItems: PropTypes.array.isRequired,
    toggleSelectedItems: PropTypes.func.isRequired,
  };

  render() {
    const {
      onClickClearSilences,
      onClickDelete,
      onClickExecute,
      onClickSilence,
      onChangeQuery,
      selectedItems,
      toggleSelectedItems,
      rowCount,
    } = this.props;

    const selectedCount = selectedItems.length;
    const selectedSilenced = selectedItems.filter(ch => !ch.silences.length);

    return (
      <ListHeader
        sticky
        selectedCount={selectedCount}
        rowCount={rowCount}
        onClickSelect={toggleSelectedItems}
        renderBulkActions={() => (
          <CollapsingMenu>
            <CollapsingMenu.Button
              title="Execute"
              icon={<QueueIcon />}
              onClick={onClickExecute}
              alt="Queue an adhoc execution of the selected checks."
              pinned
            />
            <CollapsingMenu.Button
              alt="Create a silence targeting selected checks."
              disabled={selectedSilenced.length === 0}
              icon={<SilenceIcon />}
              onClick={onClickSilence}
              title="Silence"
            />
            <CollapsingMenu.Button
              title="Unsilence"
              icon={<UnsilenceIcon />}
              onClick={onClickClearSilences}
              alt="Clear silences associated with selected checks."
              disabled={selectedSilenced.length > 0}
            />
            <ConfirmDelete
              onSubmit={ev => {
                onClickDelete(ev);
                close();
              }}
            >
              {confirm => (
                <CollapsingMenu.Button
                  title="Delete"
                  icon={<DeleteIcon />}
                  onClick={confirm.open}
                />
              )}
            </ConfirmDelete>
          </CollapsingMenu>
        )}
        renderActions={() => (
          <CollapsingMenu>
            <CollapsingMenu.SubMenu
              title="Sort"
              pinned
              renderMenu={({ anchorEl, close }) => (
                <ListSortMenu
                  anchorEl={anchorEl}
                  onClose={close}
                  options={["NAME"]}
                  onChangeQuery={onChangeQuery}
                />
              )}
            />
          </CollapsingMenu>
        )}
      />
    );
  }
}

export default ChecksListHeader;
