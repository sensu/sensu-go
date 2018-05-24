import React from "react";
import PropTypes from "prop-types";
import warning from "warning";

class CollapsingMenuItemBase extends React.PureComponent {
  static propTypes = {
    renderMenuItem: PropTypes.node.isRequired,
    renderButton: PropTypes.node.isRequired,
    renderAs: PropTypes.oneOf(["button", "menu-item", null]),
  };

  static defaultProps = {
    renderAs: null,
  };

  render() {
    const { renderMenuItem, renderButton, renderAs } = this.props;
    warning(
      `
      CollapsingMenu.MenuItem component's renderAs prop was not set. This has
      likely occurred because the component was used outside of the scope of the
      CollapsingMenu component.
    `,
      renderAs === null,
    );

    return renderAs === "menu-item" ? renderMenuItem : renderButton;
  }
}

export default CollapsingMenuItemBase;
