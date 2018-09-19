import React from "react";
import PropTypes from "prop-types";

import { Context } from "/components/partials/ToolbarMenu";
import AdaptiveMenuItem from "./AdaptiveMenuItem";

class MenuItem extends React.Component {
  static displayName = "ToolbarMenuItems.MenuItem.Connected";

  render() {
    return (
      <Context.Consumer>
        {({ collapsed, close }) => (
          <PureMenuItem collapsed={collapsed} close={close} {...this.props} />
        )}
      </Context.Consumer>
    );
  }
}

// eslint-disable-next-line react/no-multi-comp
class PureMenuItem extends React.PureComponent {
  static displayName = "ToolbarMenuItems.MenuItem.Pure";

  static propTypes = {
    autoClose: PropTypes.bool,
    onClick: PropTypes.func,
  };

  static defaultProps = {
    autoClose: true,
    onClick: () => null,
  };

  handleClick = ev => {
    if (this.props.autoClose) {
      close(ev);
    }
    this.props.onClick(ev);
  };

  render() {
    return <AdaptiveMenuItem {...this.props} onClick={this.handleClick} />;
  }
}

export default MenuItem;
