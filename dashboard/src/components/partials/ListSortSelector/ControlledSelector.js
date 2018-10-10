import React from "react";
import PropTypes from "prop-types";

import MenuController from "/components/controller/MenuController";
import Disclosure from "/components/partials/ToolbarMenuItems/Disclosure";
import RootRef from "@material-ui/core/RootRef";

import Menu from "./Menu";

class ListSortSelector extends React.Component {
  static propTypes = {
    renderButton: PropTypes.func,
  };

  static defaultProps = {
    renderButton: ({ idx, open, ref }) => (
      <RootRef rootRef={ref}>
        <Disclosure
          aria-owns={idx}
          aria-haspopup="true"
          label="Sort"
          onClick={open}
        />
      </RootRef>
    ),
  };

  renderMenu = ({ anchorEl, idx, close }) => {
    const { renderButton, ...props } = this.props;
    return <Menu anchorEl={anchorEl} id={idx} onClose={close} {...props} />;
  };

  render() {
    return (
      <MenuController renderMenu={this.renderMenu}>
        {this.props.renderButton}
      </MenuController>
    );
  }
}

export default ListSortSelector;
