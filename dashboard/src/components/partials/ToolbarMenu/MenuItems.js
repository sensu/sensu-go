import React from "react";
import PropTypes from "prop-types";

import ButtonSet from "/components/ButtonSet";
import MoreMenu from "/components/partials/MoreMenu";

class MenuItems extends React.PureComponent {
  static displayName = "ToolbarMenu.Items";

  static propTypes = {
    items: PropTypes.node,
    itemsRef: PropTypes.object,
    overflow: PropTypes.func,
  };

  static defaultProps = {
    items: null,
    itemsRef: {},
    overflow: null,
  };

  render() {
    const { items, itemsRef, overflow } = this.props;

    return (
      <React.Fragment>
        <ButtonSet ref={itemsRef}>{items}</ButtonSet>
        {overflow && <MoreMenu renderMenu={overflow} />}
      </React.Fragment>
    );
  }
}

export default MenuItems;
