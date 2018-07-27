import React from "react";
import PropTypes from "prop-types";

import withWidth, { isWidthUp } from "@material-ui/core/withWidth";
import ButtonSet from "/components/ButtonSet";
import Menu from "@material-ui/core/Menu";
import VerticalDisclosureButton from "/components/VerticalDisclosureButton";

import Item from "./Item";
import Button from "./Button";
import SubMenu from "./SubMenu";

const Context = React.createContext();

let id = 0;
const getNextId = () => {
  id += 1;
  return id;
};

const breakpoints = ["xs", "sm", "md", "lg", "xl"];

class CollapsingMenu extends React.PureComponent {
  static propTypes = {
    breakpoint: PropTypes.oneOf(breakpoints.slice(1)),
    children: PropTypes.node.isRequired,
    width: PropTypes.string.isRequired,
  };

  static defaultProps = {
    breakpoint: "sm",
  };

  static Item = Item;
  static Button = Button;
  static SubMenu = SubMenu;

  constructor(props) {
    super(props);
    this._id = getNextId();
  }

  state = {
    anchorEl: null,
  };

  renderExpanded = () => {
    const { children } = this.props;

    return (
      <ButtonSet>
        <Context.Provider
          value={{
            collapsed: false,
            parent: "buttonset",
            close,
          }}
        >
          {children}
        </Context.Provider>
      </ButtonSet>
    );
  };

  renderCollapsed = () => {
    const { children } = this.props;

    const menuId = `collapsed-menu-${this._id}`;
    const close = () => this.setState({ anchorEl: null });

    return (
      <React.Fragment>
        <ButtonSet>
          <Context.Provider
            value={{
              collapsed: true,
              parent: "buttonset",
              close,
            }}
          >
            {children}
          </Context.Provider>
        </ButtonSet>

        <VerticalDisclosureButton
          aria-label="More"
          aria-owns={menuId}
          aria-haspopup="true"
          onClick={ev => this.setState({ anchorEl: ev.currentTarget })}
        />
        <Menu
          id={menuId}
          anchorEl={this.state.anchorEl}
          open={Boolean(this.state.anchorEl)}
          onClose={close}
          keepMounted
        >
          <Context.Provider
            value={{
              collapsed: true,
              parent: "menu",
              close,
            }}
          >
            {children}
          </Context.Provider>
        </Menu>
      </React.Fragment>
    );
  };

  render() {
    const { breakpoint, width } = this.props;

    if (isWidthUp(breakpoint, width)) {
      return this.renderExpanded();
    }
    return this.renderCollapsed();
  }
}

const EnhancedComponent = withWidth()(CollapsingMenu);
export { EnhancedComponent as default, Context };
