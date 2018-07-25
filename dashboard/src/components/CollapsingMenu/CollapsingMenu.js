import React from "react";
import PropTypes from "prop-types";
import Hidden from "@material-ui/core/Hidden";
import Menu from "@material-ui/core/Menu";
import ButtonSet from "/components/ButtonSet";
import VerticalDisclosureButton from "/components/VerticalDisclosureButton";
import MenuItem from "./MenuItem";
import Button from "./Button";

const Context = React.createContext();

let id = 0;
const getNextId = () => {
  id += 1;
  return id;
};

const breakpoints = ["xs", "sm", "md", "lg", "xl"];

class CollapsingMenu extends React.PureComponent {
  static propTypes = {
    children: PropTypes.node.isRequired,
    breakpoint: PropTypes.oneOf(breakpoints.slice(1)),
  };

  static defaultProps = {
    breakpoint: "sm",
  };

  static MenuItem = MenuItem;
  static Button = Button;

  constructor(props) {
    super(props);
    this._id = getNextId();
  }

  state = {
    anchorEl: null,
  };

  render() {
    const { breakpoint, children } = this.props;

    const menuId = `collapsed-menu-${this._id}`;
    const close = () => this.setState({ anchorEl: null });

    // find the previous breakpoint; clamp if we are already on the boundary.
    const breakpointIdx = breakpoints.indexOf(breakpoint);
    const prevBreakpoint = breakpoints[breakpointIdx ? breakpointIdx - 1 : 0];

    return (
      <React.Fragment>
        <Hidden {...{ [`${prevBreakpoint}Down`]: true }}>
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
        </Hidden>
        <Hidden {...{ [`${breakpoint}Up`]: true }}>
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
        </Hidden>
      </React.Fragment>
    );
  }
}

export { CollapsingMenu as default, Context };
