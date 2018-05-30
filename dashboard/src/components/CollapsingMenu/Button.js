import React from "react";
import PropTypes from "prop-types";
import Button from "@material-ui/core/Button";
import MenuItem from "@material-ui/core/MenuItem";
import ListItemIcon from "@material-ui/core/ListItemIcon";
import ListItemText from "@material-ui/core/ListItemText";
import ButtonIcon from "/components/ButtonIcon";
import Item from "./MenuItem";

class CollapsingButton extends React.PureComponent {
  static displayName = "CollapsingMenu.Button";

  static propTypes = {
    icon: PropTypes.node,
    title: PropTypes.string.isRequired,
    subtitle: PropTypes.string,
    onClick: PropTypes.func.isRequired,
  };

  static defaultProps = {
    icon: null,
    subtitle: null,
  };

  render() {
    const { icon, title, subtitle, onClick: onClickProp } = this.props;

    return (
      <Item>
        {({ collapsed, close }) => {
          const onClick = ev => {
            if (onClickProp) {
              onClickProp(ev);
            }
            close(ev);
          };

          if (collapsed) {
            return (
              <MenuItem onClick={onClick}>
                <ListItemIcon>{icon}</ListItemIcon>
                <ListItemText inset primary={title} secondary={subtitle} />
              </MenuItem>
            );
          }

          return (
            <Button onClick={onClick}>
              <ButtonIcon>{icon}</ButtonIcon>
              {title}
            </Button>
          );
        }}
      </Item>
    );
  }
}

export default CollapsingButton;
