import React from "react";
import PropTypes from "prop-types";
import BaseMenu from "@material-ui/core/Menu";

class Menu extends React.Component {
  static displayName = "ToolbarSelect.Menu";

  static propTypes = {
    onClose: PropTypes.func.isRequired,
    children: PropTypes.node.isRequired,
    anchorEl: PropTypes.object.isRequired,
    onChange: PropTypes.func.isRequired,
  };

  render() {
    const { anchorEl, children, onClose, onChange } = this.props;

    return (
      <BaseMenu open onClose={onClose} anchorEl={anchorEl}>
        {React.Children.map(children, child => {
          const onClick = event => {
            if (child.props.onClick) {
              child.props.onClick(event);
              if (event.defaultPrevented) {
                return;
              }
            }

            if (child.props.value !== undefined) {
              onChange(child.props.value);
            }

            onClose();
          };

          return React.cloneElement(child, { onClick });
        })}
      </BaseMenu>
    );
  }
}

export default Menu;
