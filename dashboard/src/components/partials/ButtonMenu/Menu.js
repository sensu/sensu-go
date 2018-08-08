import React from "react";
import PropTypes from "prop-types";
import Menu from "@material-ui/core/Menu";

class ButtonMenu extends React.Component {
  static propTypes = {
    onClose: PropTypes.func.isRequired,
    children: PropTypes.node.isRequired,
    anchorEl: PropTypes.string.isRequired,
    onChange: PropTypes.func.isRequired,
  };

  render() {
    const { anchorEl, children, onClose, onChange } = this.props;

    return (
      <Menu open onClose={onClose} anchorEl={anchorEl}>
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
      </Menu>
    );
  }
}

export default ButtonMenu;
