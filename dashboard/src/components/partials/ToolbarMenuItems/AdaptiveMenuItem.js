import React from "react";
import PropTypes from "prop-types";

import Button from "./Button";
import CollapsedItem from "./CollapsedItem";

class MenuItem extends React.PureComponent {
  static displayName = "ToolbarMenuItems.MenuItem";

  static propTypes = {
    collapsed: PropTypes.bool,
    color: PropTypes.string,
    description: PropTypes.node,
    icon: PropTypes.node,
    onClick: PropTypes.func,
    title: PropTypes.node,
    titleCondensed: PropTypes.node,
  };

  static defaultProps = {
    collapsed: false,
    color: "inherit",
    description: null,
    icon: null,
    onClick: () => null,
    title: null,
    titleCondensed: null,
  };

  render() {
    const {
      collapsed,
      color,
      description,
      icon,
      onClick,
      title,
      titleCondensed,
      ...props
    } = this.props;

    if (collapsed) {
      return (
        <CollapsedItem
          icon={icon}
          onClick={onClick}
          primary={title}
          {...props}
        />
      );
    }

    return (
      <Button
        description={description}
        icon={icon}
        label={titleCondensed || title}
        onClick={onClick}
        color={color}
        {...props}
      />
    );
  }
}

export default MenuItem;
