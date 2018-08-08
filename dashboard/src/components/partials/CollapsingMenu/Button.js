import React from "react";
import PropTypes from "prop-types";
import BaseButton from "@material-ui/core/Button";
import ButtonIcon from "/components/ButtonIcon";
import ListItemIcon from "@material-ui/core/ListItemIcon";
import ListItemText from "@material-ui/core/ListItemText";
import MenuItem from "@material-ui/core/MenuItem";
import Tooltip from "@material-ui/core/Tooltip";

import Item from "./Item";

const DescribedButton = ({ alt, ...props }) => {
  if (alt) {
    return (
      <Tooltip title={alt}>
        <BaseButton {...props} />
      </Tooltip>
    );
  }
  return <BaseButton {...props} />;
};
DescribedButton.propTypes = { alt: PropTypes.string };
DescribedButton.defaultProps = { alt: null };

class Button extends React.PureComponent {
  static displayName = "CollapsingMenu.Button";

  static propTypes = {
    alt: PropTypes.string,
    disabled: PropTypes.bool,
    icon: PropTypes.node,
    title: PropTypes.string.isRequired,
    subtitle: PropTypes.string,
    onClick: PropTypes.func.isRequired,
    pinned: PropTypes.bool,
  };

  static defaultProps = {
    alt: null,
    color: "inherit",
    disabled: false,
    icon: null,
    subtitle: null,
    pinned: false,
  };

  render() {
    const {
      icon,
      title,
      subtitle,
      onClick: onClickProp,
      pinned,
      ...props
    } = this.props;

    const menuProps = {
      disabled: props.disabled,
    };
    const buttonProps = {
      alt: props.alt,
      color: props.color,
      disabled: props.disabled,
    };

    return (
      <Item
        pinned={pinned}
        renderMenuItem={({ close }) => {
          const onClick = ev => {
            if (onClickProp) {
              onClickProp(ev);
            }
            close(ev); // TODO: Don't automatically handle this.
          };

          return (
            <MenuItem onClick={onClick} {...menuProps}>
              <ListItemIcon>{icon}</ListItemIcon>
              <ListItemText inset primary={title} secondary={subtitle} />
            </MenuItem>
          );
        }}
        renderToolbarItem={() => (
          <DescribedButton onClick={onClickProp} {...buttonProps}>
            <ButtonIcon>{icon}</ButtonIcon>
            {title}
          </DescribedButton>
        )}
      />
    );
  }
}

export default Button;
