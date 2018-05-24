import React from "react";
import PropTypes from "prop-types";
import Hidden from "@material-ui/core/Hidden";
import Menu from "@material-ui/core/Menu";
import ButtonSet from "/components/ButtonSet";
import VerticalDisclosureButton from "/components/VerticalDisclosureButton";

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

  constructor(props) {
    super(props);
    this._id = getNextId();
  }

  state = {
    anchorEl: null,
  };

  render() {
    const { breakpoint, children } = this.props;

    const buttons = React.Children.map(children, child =>
      React.cloneElement(child, { renderAs: "button" }),
    );
    const menuId = `collapsed-menu-${this._id}`;
    const menuItems = React.Children.map(children, child =>
      React.cloneElement(child, {
        renderAs: "menu-item",
        onClick: ev => {
          if (child.props.onClick) child.props.onClick(ev);
          this.setState({ anchorEl: null });
        },
      }),
    );

    const breakpointIdx = breakpoints.indexOf(breakpoint);
    const prevBreakpoint = breakpoints[breakpointIdx ? breakpointIdx - 1 : 0];

    return (
      <React.Fragment>
        <Hidden {...{ [`${prevBreakpoint}Down`]: true }}>
          <ButtonSet>{buttons}</ButtonSet>
        </Hidden>
        <Hidden {...{ [`${breakpoint}Up`]: true }}>
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
            onClose={() => this.setState({ anchorEl: null })}
          >
            {menuItems}
          </Menu>
        </Hidden>
      </React.Fragment>
    );
  }
}

export default CollapsingMenu;
