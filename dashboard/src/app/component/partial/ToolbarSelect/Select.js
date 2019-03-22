import React from "react";
import PropTypes from "prop-types";

import Button from "./Button";
import Controller from "./Controller";

class ToolbarSelect extends React.Component {
  static propTypes = {
    button: PropTypes.node,
    children: PropTypes.arrayOf(PropTypes.node).isRequired,
    onChange: PropTypes.func.isRequired,
    title: PropTypes.node,
  };

  static defaultProps = {
    button: null,
    title: null,
  };

  render() {
    const {
      button: buttonProp,
      children,
      onChange,
      title,
      ...props
    } = this.props;

    let button;
    if (buttonProp) {
      button = React.cloneElement(buttonProp, { ...props, title });
    } else {
      button = <Button title={title} {...props} />;
    }

    return (
      <Controller options={children} onChange={onChange}>
        {ctrl => React.cloneElement(button, { onClick: ctrl.open })}
      </Controller>
    );
  }
}

export default ToolbarSelect;
