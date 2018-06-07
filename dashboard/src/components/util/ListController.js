import React from "react";
import PropTypes from "prop-types";

const setKeySelected = (key, keySelected) => state => {
  if (
    keySelected &&
    // Prevent adding duplicated keys to the selectedKeys array.
    !state.selectedKeys.includes(key) &&
    // Prevent including a key that is not in the current items array.
    state.keys.includes(key)
  ) {
    const selectedKeys = state.selectedKeys.concat([key]);
    return { selectedKeys };
  }

  if (
    !keySelected &&
    // Prevent unnecessary state updates if the key is not selected.
    state.selectedKeys.includes(key)
  ) {
    const selectedKeys = state.selectedKeys.filter(
      selectedKey => key !== selectedKey,
    );
    return { selectedKeys };
  }

  return null;
};

class ListController extends React.PureComponent {
  static defaultProps = {
    renderItem: PropTypes.func.isRequired,
    renderEmptyState: PropTypes.func.isRequired,
    children: PropTypes.func.isRequired,
  };

  state = {
    selectedKeys: [],
    items: [],
    keys: [],
    getItemKey: undefined,
    renderItem: undefined,
    render: undefined,
    renderEmptyState: undefined,
  };

  static getDerivedStateFromProps(props, previousState) {
    const {
      renderItem,
      renderEmptyState,
      getItemKey,
      items,
      children: render,
    } = props;

    let state = previousState;

    if (state.items !== items || state.getItemKey !== getItemKey) {
      const keys = props.items.map(item => getItemKey(item));
      const selectedKeys = state.selectedKeys.filter(key => keys.includes(key));
      state = { ...state, selectedKeys, items, keys, getItemKey };
    }

    if (state.renderItem !== renderItem) {
      state = { ...state, renderItem };
    }

    if (state.render !== render) {
      state = { ...state, render };
    }

    if (state.renderEmptyState !== renderEmptyState) {
      state = { ...state, renderEmptyState };
    }

    if (state === previousState) {
      return null;
    }

    return state;
  }

  setItemSelected = (item, itemSelected) => {
    this.setState(state => {
      const key = state.getItemKey(item);
      return setKeySelected(key, itemSelected)(state);
    });
  };

  setKeySelected = (key, keySelected) => {
    this.setState(setKeySelected(key, keySelected));
  };

  setSelectedKeys = selectedKeys => {
    this.setState(state => ({
      selectedKeys: state.keys.filter(key => selectedKeys.includes(key)),
    }));
  };

  render() {
    const {
      items,
      keys,
      selectedKeys,
      renderItem,
      render,
      renderEmptyState,
    } = this.state;

    return render({
      children: items.length
        ? items.map((item, i) => {
            const key = keys[i];
            const selected = selectedKeys.includes(key);

            return renderItem({
              key,
              item,
              selected,
              setSelected: keySelected => this.setKeySelected(key, keySelected),
              toggleSelected: () => this.setKeySelected(key, !selected),
            });
          })
        : renderEmptyState(),
      keys,
      selectedKeys,
      selectedItems: items.filter((item, i) => {
        const key = keys[i];
        return selectedKeys.includes(key);
      }),
      setKeySelected: this.setKeySelected,
      setItemSelected: this.setItemSelected,
      setSelectedKeys: this.setSelectedKeys,
      toggleSelectedItems: () =>
        selectedKeys.length === keys.length
          ? this.setSelectedKeys([])
          : this.setSelectedKeys(keys),
    });
  }
}

export default ListController;
