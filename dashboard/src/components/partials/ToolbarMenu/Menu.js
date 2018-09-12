import React from "react";
import ReactDOM from "react-dom";
import PropTypes from "prop-types";
import EventListener from "react-event-listener";

import MenuItems from "./MenuItems";
import MenuItem from "./MenuItem";

const Context = React.createContext();

class ToolbarMenu extends React.PureComponent {
  static propTypes = {
    children: PropTypes.oneOfType([
      PropTypes.arrayOf(PropTypes.node),
      PropTypes.node,
    ]).isRequired,
  };

  static Item = MenuItem;

  state = {
    toolbarWidth: null,
    buttonWidths: null,
  };

  static getDerivedStateFromProps(_, state) {
    return { ...state, buttonWidths: null };
  }

  componentDidMount() {
    this.updateWidths();
  }

  componentDidUpdate() {
    this.updateWidths();
  }

  updateWidths = () => {
    if (this.state.buttonWidths !== null) {
      return;
    }

    // eslint-disable-next-line react/no-find-dom-node
    const el = ReactDOM.findDOMNode(this.barRef.current);
    const toolbarWidth = el.getBoundingClientRect().width;

    const children = React.Children.map(this.props.children, child => child);
    const widths = [];

    for (let [i, j] = [0, 0]; i < children.length; i += 1) {
      const child = children[i];
      if (child.props.visible === "never") {
        continue; // eslint-disable-line no-continue
      }

      const childEl = el.children.item(j);
      j += 1;

      const rect = childEl.getBoundingClientRect();
      const styles = window.getComputedStyle(childEl);

      const marginLeft = parseFloat(styles.marginLeft, 10);
      const marginRight = parseFloat(styles.marginRight, 10);

      widths[i] = rect.width + marginLeft + marginRight;
    }

    if (widths.length === 0) {
      return;
    }

    // eslint-disable-next-line react/no-did-update-set-state
    this.setState({ buttonWidths: widths, toolbarWidth });
  };

  barRef = React.createRef();

  handleResize = () => {
    this.setState({ buttonWidths: null, toolbarWidth: null });
  };

  renderItems = items => {
    const childProps = {
      close: () => null,
      collapsed: false,
    };

    return React.Children.map(items, child => {
      let children = child.props.children;
      if (typeof children === "function") {
        children = children(childProps);
      }

      return (
        <Context.Provider value={childProps}>
          {React.cloneElement(child, { children })}
        </Context.Provider>
      );
    });
  };

  renderOverflow = items => {
    if (items.length === 0 && this.state.buttonWidths !== null) {
      return null;
    }

    return renderProps =>
      React.Children.map(items, child => {
        const childProps = {
          collapsed: true,
          ...renderProps,
        };

        let children = child.props.children;
        if (typeof children === "function") {
          children = children(childProps);
        }

        return (
          <Context.Provider value={childProps}>
            {React.cloneElement(child, { children })}
          </Context.Provider>
        );
      });
  };

  render() {
    const { children: childrenProp } = this.props;
    const { toolbarWidth, buttonWidths: buttonWidthsState } = this.state;

    const buttonWidths = buttonWidthsState || [];
    const children = React.Children.map(childrenProp, child => child);

    let remainingWidth = children.reduce((acc, child, i) => {
      if (child.props.visible === "always") {
        return acc - buttonWidths[i];
      }
      return acc;
    }, toolbarWidth);

    let visible = [];
    let collapsed = [];

    for (let i = 0; i < children.length; i += 1) {
      const item = children[i];
      const itemWidth = buttonWidths[i];
      const visibility = item.props.visible;

      if (item.type !== MenuItem) {
        throw new Error(
          "A partitioned toolbar's children must be of type ToolbarMenu.Item",
        );
      }

      if (visibility === "never") {
        collapsed = [...collapsed, item];
      } else if (buttonWidthsState === null || visibility === "always") {
        visible = [...visible, item];
      } else if (remainingWidth >= itemWidth) {
        remainingWidth -= itemWidth;
        visible = [...visible, item];
      } else {
        collapsed = [...collapsed, item];
      }
    }

    return (
      <React.Fragment>
        <EventListener target="window" onResize={this.handleResize} />

        <MenuItems
          items={this.renderItems(visible)}
          itemsRef={this.barRef}
          overflow={this.renderOverflow(collapsed)}
        />
      </React.Fragment>
    );
  }
}

export { Context, ToolbarMenu as default };
