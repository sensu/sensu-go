import React from "react";
import PropTypes from "prop-types";

import BaseButton from "@material-ui/core/Button";
import ButtonIcon from "/components/ButtonIcon";
import Tooltip from "@material-ui/core/Tooltip";

class Button extends React.Component {
  static displayName = "ToolbarMenuItems.Button";

  static propTypes = {
    color: PropTypes.string,
    component: PropTypes.func,
    description: PropTypes.node,
    disabled: PropTypes.bool,
    icon: PropTypes.node,
    iconAlignment: PropTypes.oneOf(["left", "right"]),
    iconRef: PropTypes.func,
    label: PropTypes.node.isRequired,
    ornament: PropTypes.node,
    ornamentRef: PropTypes.object,
  };

  static defaultProps = {
    color: "inherit",
    component: BaseButton,
    description: null,
    disabled: false,
    icon: null,
    iconAlignment: "left",
    iconRef: null,
    ornament: null,
    ornamentRef: null,
  };

  render() {
    const {
      color,
      component: Component,
      description,
      icon: iconProp,
      iconAlignment,
      iconRef,
      label,
      ornament: ornamentProp,
      ornamentRef,
      ...props
    } = this.props;

    let icon;
    if (iconProp) {
      icon = (
        <ButtonIcon alignment={iconAlignment} ref={iconRef}>
          {iconProp}
        </ButtonIcon>
      );
    }

    let ornament;
    if (ornamentProp) {
      ornament = (
        <ButtonIcon alignment="right" ref={ornamentRef}>
          {ornamentProp}
        </ButtonIcon>
      );
    }

    const button = (
      <Component aria-label={label} color={color} {...props}>
        {iconAlignment === "left" && icon}
        {label}
        {iconAlignment === "right" && icon}
        {ornament}
      </Component>
    );

    if (description && !props.disabled) {
      return <Tooltip title={description}>{button}</Tooltip>;
    }
    return button;
  }
}

export default Button;
