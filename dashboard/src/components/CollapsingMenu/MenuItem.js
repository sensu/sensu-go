import React from "react";
import PropTypes from "prop-types";
import { Context } from "./CollapsingMenu";

class MenuItem extends React.PureComponent {
  static displayName = "CollapsingMenu.MenuItem";

  static propTypes = {
    children: PropTypes.func.isRequired,
  };

  render() {
    return <Context.Consumer>{this.props.children}</Context.Consumer>;
  }
}

export default MenuItem;
