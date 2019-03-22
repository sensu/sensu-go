import React from "react";
import PropTypes from "prop-types";
import { shallowEqual } from "/lib/util/array";

import MenuItem from "./Item";

// Return given set of menu items and a width return map describing which items
// are visible and which items should be collapsed.
const partitionItems = ({ items, itemWidths, width: widthProp }) => {
  const width = widthProp !== null ? widthProp : 4096;

  // 1. We first want initialize the remaining width accounting for any items
  //    that are /always/ displayed.
  let remainingWidth = items.reduce((acc, item) => {
    if (item.props.visible === "always") {
      return acc - (itemWidths[item.props.id] || 0);
    }
    return acc;
  }, width);

  // 2. On our second pass we ensure that we collect each item based on their
  //    configured visibility and the remaining room.
  const visible = [];
  const collapsed = [];
  items.forEach(item => {
    const id = item.props.id;
    const itemWidth = itemWidths[id] || 0;
    const visibility = item.props.visible;

    if (visibility === "always") {
      visible.push(id);
    } else if (visibility === "if-room" && remainingWidth >= itemWidth) {
      remainingWidth -= itemWidth;
      visible.push(id);
    } else {
      collapsed.push(id);
    }
  });

  return { visible, collapsed };
};

class Partitioner extends React.Component {
  static displayName = "ToolbarMenu.Partitioner";

  static propTypes = {
    children: PropTypes.func.isRequired,
    items: PropTypes.node.isRequired,
    width: PropTypes.number, // eslint-disable-line react/no-unused-prop-types
  };

  static defaultProps = {
    width: null,
  };

  static getDerivedStateFromProps(props, state) {
    if (props.width === state.width && props.items === state.items) {
      return null;
    }

    // Ensure children are menu items
    if (process.env.NODE_ENV !== "production") {
      React.Children.toArray(props.items).forEach(child => {
        if (child.type !== MenuItem) {
          throw new Error(
            "A partitioned toolbar's children must be of type ToolbarMenu.Item",
          );
        }
      });
    }

    const nextState = {
      ...state,
      items: React.Children.toArray(props.items),
      width: props.width,
    };

    const partition = partitionItems(nextState);
    return { ...nextState, ...partition };
  }

  state = {
    items: [],
    itemWidths: {},
    width: null,
    visible: [],
    collapsed: [],
  };

  shouldComponentUpdate(props, state) {
    const { visible: aVisible, collapsed: aCollapsed } = this.state;
    const { visible: bVisible, collapsed: bCollapsed } = state;

    return (
      this.props.items !== props.items ||
      !shallowEqual(aVisible, bVisible) ||
      !shallowEqual(aCollapsed, bCollapsed)
    );
  }

  handleItemResize = id => rect => {
    this.setState(state => {
      if (state.itemWidths[id] === rect.width) {
        return null;
      }

      const itemWidths = { [id]: rect.width, ...state.itemWidths };
      const { visible, collapsed } = partitionItems({ ...state, itemWidths });
      return { itemWidths, visible, collapsed };
    });
  };

  render() {
    const visible = [];
    const collapsed = [];

    this.state.items.forEach(item => {
      const id = item.props.id;

      if (this.state.visible.includes(id)) {
        visible.push(
          React.cloneElement(item, { onResize: this.handleItemResize(id) }),
        );
      }
      if (this.state.collapsed.includes(id)) {
        collapsed.push(item);
      }
    });

    return this.props.children({
      visible,
      collapsed,
    });
  }
}

export default Partitioner;
