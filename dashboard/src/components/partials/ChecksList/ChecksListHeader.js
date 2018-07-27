import React from "react";
import PropTypes from "prop-types";

import Button from "@material-ui/core/Button";
import ButtonSet from "/components/ButtonSet";
import ConfirmDelete from "/components/partials/ConfirmDelete";
import DropdownArrow from "@material-ui/icons/ArrowDropDown";
import IconButton from "/components/partials/IconButton";
import ListHeader from "/components/partials/ListHeader";
import ListSortMenu from "/components/partials/ListSortMenu";
import MenuController from "/components/controller/MenuController";
import RootRef from "@material-ui/core/RootRef";
import Typography from "@material-ui/core/Typography";

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
          <ButtonSet>
            {selectedSilenced.length > 0 && (
              <Button onClick={onClickSilence}>
                <Typography variant="button">Silence</Typography>
              </Button>
            )}
            {selectedSilenced.length === 0 && (
              <Button onClick={onClickClearSilences}>
                <Typography variant="button">Unsilence</Typography>
              </Button>
            )}
            <Button onClick={onClickExecute}>
              <Typography variant="button">Execute</Typography>
            </Button>
            <ConfirmDelete
              onSubmit={ev => {
                onClickDelete(ev);
                close();
              }}
            >
              {confirm => (
                <Button onClick={confirm.open}>
                  <Typography variant="button">Delete</Typography>
                </Button>
              )}
            </ConfirmDelete>
          </ButtonSet>
        )}
        renderActions={() => (
          <ButtonSet>
            <MenuController
              renderMenu={({ anchorEl, close }) => (
                <ListSortMenu
                  anchorEl={anchorEl}
                  onClose={close}
                  options={["NAME"]}
                  onChangeQuery={onChangeQuery}
                />
              )}
            >
              {({ open, ref }) => (
                <RootRef rootRef={ref}>
                  <IconButton onClick={open} icon={<DropdownArrow />}>
                    Sort
                  </IconButton>
                </RootRef>
              )}
            </MenuController>
          </ButtonSet>
        )}
      />
    );
  }
}

export default ChecksListHeader;
