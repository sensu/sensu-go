import React from "react";
import PropTypes from "prop-types";
import EventListener from "react-event-listener";
import ResizeObserver from "react-resize-observer";
import debounce from "debounce";
import { shallowEqual } from "/utils/array";

import RootRef from "@material-ui/core/RootRef";
import MenuController from "/components/controller/MenuController";

import Menu from "./OverflowMenu";
import MenuButton from "./OverflowButton";
import MenuItem from "./Item";
import Partitioner from "./Partitioner";
import Autosizer from "./Autosizer";

const Context = React.createContext({});

// Resize events are handled syncronously and can cause significant thrashing
// unless debounced.
const windowResizeInterval = 200;

class ToolbarMenu extends React.PureComponent {
  static propTypes = {
    children: PropTypes.oneOfType([
      PropTypes.arrayOf(PropTypes.node),
      PropTypes.node,
    ]).isRequired,
    width: PropTypes.number,
  };

  static defaultProps = {
    width: null,
  };

  static Autosizer = Autosizer;
  static Item = MenuItem;

  // If the menu items change poison the buttons container's width, to ensure
  // that we are displaying as many buttons as possible.
  static getDerivedStateFromProps(props, state) {
    const ids = React.Children.map(props.children, child => child.props.id);
    const visibility = React.Children.map(
      props.children,
      child => child.props.visible,
    );
    if (
      // if we add more props that this equation depends on
      // we should really just rewrite shallowEqual and send these
      // in a single array, but for now, it's fine
      !shallowEqual(ids, state.ids) ||
      !shallowEqual(visibility, state.visibility)
    ) {
      return { ids, visibility, buttonsWidth: null };
    }
    return null;
  }

  state = {
    // List of item ids
    ids: [],
    visibility: [],

    // Width of buttons container
    buttonsWidth: null,

    // Assume the overflow button is default size of icon button.
    overflowButtonWidth: 48,
  };

  componentWillUnmount() {
    this.updateButtonsWidth.clear();
  }

  handleOverflowButtonResize = rect => {
    this.setState(state => {
      if (state.overflowButtonWidth === rect.width) {
        return null;
      }

      return { overflowButtonWidth: rect.width };
    });
  };

  handleButtonsResize = rect => {
    this.setState(state => {
      if (state.buttonsWidth === rect.width) {
        return null;
      }

      return { buttonsWidth: rect.width };
    });
  };

  handleWindowResize = event => {
    this.updateButtonsWidth(event.currentTarget);
  };

  updateButtonsWidth = debounce(currentTarget => {
    const newWidth = currentTarget.innerWidth;
    const oldWidth = this.windowWidth || 0;

    // If the window grew in size and the toolbar menu isn't configured to fill
    // the entire space we try rendering all the items again.
    if (!this.props.width && newWidth > oldWidth) {
      this.setState({ buttonsWidth: null }); // synchronous
    }

    this.windowWidth = newWidth;
  }, windowResizeInterval);

  buttonsWidth = () => {
    const { width } = this.props;
    const { buttonsWidth, overflowButtonWidth } = this.state;

    return width === null ? buttonsWidth : width - overflowButtonWidth;
  };

  renderButtonSet = items => {
    const ctx = { collapsed: false, close: () => null };
    const buttons = React.Children.map(items, child => (
      <Context.Provider value={ctx}>{child}</Context.Provider>
    ));

    if (this.props.width === null) {
      return (
        <div style={{ position: "relative" }}>
          <ResizeObserver onResize={this.handleButtonsResize} />
          {buttons}
        </div>
      );
    }

    return buttons;
  };

  renderOverflow = items => {
    if (items.length === 0) {
      return null;
    }

    return (
      <MenuController
        renderMenu={({ anchorEl, idx, close }) => (
          <Menu id={idx} open onClose={close} anchorEl={anchorEl}>
            {React.Children.map(items, child => (
              <Context.Provider value={{ collapsed: true, close }}>
                {child}
              </Context.Provider>
            ))}
          </Menu>
        )}
      >
        {this.renderOverflowButton}
      </MenuController>
    );
  };

  renderOverflowButton = ({ idx, isOpen, open, ref }) => {
    const button = (
      <RootRef rootRef={ref}>
        <MenuButton active={isOpen} idx={idx} onClick={open} />
      </RootRef>
    );

    if (this.props.width !== null) {
      return (
        <div style={{ display: "inline", position: "relative" }}>
          <ResizeObserver onResize={this.handleOverflowButtonResize} />
          {button}
        </div>
      );
    }
    return button;
  };

  renderItems = partition => (
    <React.Fragment>
      {this.renderButtonSet(partition.visible)}
      {this.renderOverflow(partition.collapsed)}
    </React.Fragment>
  );

  render() {
    const items = (
      <Partitioner items={this.props.children} width={this.buttonsWidth()}>
        {this.renderItems}
      </Partitioner>
    );

    if (!this.props.width) {
      return (
        <React.Fragment>
          <EventListener target="window" onResize={this.handleWindowResize} />
          {items}
        </React.Fragment>
      );
    }

    return items;
  }
}

export { Context, ToolbarMenu as default };
