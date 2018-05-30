import React from "react";
import PropTypes from "prop-types";

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

    if (process.env.NODE_ENV !== "production" && !renderAs) {
      throw new Error(
        `
        CollapsingMenu.MenuItemBase component's renderAs prop was not set. This
        likely occurred because the component was used outside of the scope of
        the CollapsingMenu component.
        `,
      );
    }
    return renderAs === "menu-item" ? renderMenuItem : renderButton;
  }
}

export default CollapsingMenuItemBase;
