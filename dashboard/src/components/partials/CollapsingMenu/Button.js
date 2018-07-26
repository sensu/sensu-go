import React from "react";
import PropTypes from "prop-types";
import BaseButton from "@material-ui/core/Button";
import ButtonIcon from "/components/ButtonIcon";
import ListItemIcon from "@material-ui/core/ListItemIcon";
import ListItemText from "@material-ui/core/ListItemText";
import MenuItem from "@material-ui/core/MenuItem";
import Item from "./Item";

class Button extends React.PureComponent {
  static displayName = "CollapsingMenu.Button";

  static propTypes = {
    icon: PropTypes.node,
    title: PropTypes.string.isRequired,
    subtitle: PropTypes.string,
    onClick: PropTypes.func.isRequired,
    pinned: PropTypes.bool,
  };

  static defaultProps = {
    icon: null,
    subtitle: null,
    pinned: false,
  };

  render() {
    const { icon, title, subtitle, onClick: onClickProp, pinned } = this.props;

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
            <MenuItem onClick={onClick}>
              <ListItemIcon>{icon}</ListItemIcon>
              <ListItemText inset primary={title} secondary={subtitle} />
            </MenuItem>
          );
        }}
        renderToolbarItem={() => (
          <BaseButton onClick={onClickProp}>
            <ButtonIcon>{icon}</ButtonIcon>
            {title}
          </BaseButton>
        )}
      />
    );
  }
}

export default Button;
