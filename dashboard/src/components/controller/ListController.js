import React from "react";
import PropTypes from "prop-types";

const arrayIntersect = (a, b) => a.filter(val => b.includes(val));

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

const setSelectedKeys = selectedKeys => state => ({
  selectedKeys: arrayIntersect(state.keys, selectedKeys),
});

class ListController extends React.PureComponent {
  static propTypes = {
    // the linter isn't smart enough, we do use these
    // eslint-disable-next-line
    renderItem: PropTypes.func.isRequired,
    // eslint-disable-next-line
    renderEmptyState: PropTypes.func.isRequired,
    // eslint-disable-next-line
    children: PropTypes.func.isRequired,
    // eslint-disable-next-line
    initialSelectedKeys: PropTypes.array,
    // eslint-disable-next-line
    items: PropTypes.array,
  };

  static defaultProps = { initialSelectedKeys: [], items: [] };

  state = {
    selectedKeys: this.props.initialSelectedKeys,
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
      const selectedKeys = arrayIntersect(state.selectedKeys, keys);
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

  setSelectedItems = selectedItems => {
    this.setState(state => {
      const keys = selectedItems.map(state.getItemKey);
      return setSelectedKeys(keys)(state);
    });
  };

  setSelectedKeys = selectedKeys => {
    this.setState(setSelectedKeys(selectedKeys));
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
      setSelectedItems: this.setSelectedItems,
      toggleSelectedItems: () =>
        selectedKeys.length === 0
          ? this.setSelectedKeys(keys)
          : this.setSelectedKeys([]),
    });
  }
}

export default ListController;
