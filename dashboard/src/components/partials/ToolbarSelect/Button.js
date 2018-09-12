import React from "react";
import PropTypes from "prop-types";

import BaseButton from "@material-ui/core/Button";

class Button extends React.PureComponent {
  static displayName = "ToolbarSelect.Button";

  static propTypes = {
    title: PropTypes.string.isRequired,
    onClick: PropTypes.func.isRequired,
  };

  render() {
    const { title, onClick, ...props } = this.props;

    return (
      <BaseButton onClick={onClick} {...props}>
        {title}
      </BaseButton>
    );
  }
}

export default Button;
