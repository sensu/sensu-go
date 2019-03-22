import React from "react";
import PropTypes from "prop-types";

class HoverController extends React.PureComponent {
  static propTypes = {
    children: PropTypes.node.isRequired,
    onHover: PropTypes.func.isRequired,
  };

  onEnter = ev => {
    if (this.props.children.props.onMouseEnter) {
      this.props.children.props.onMouseEnter(ev);
    }
    this.props.onHover({ hovered: true, ...ev });
  };

  onLeave = ev => {
    if (this.props.children.props.onMouseLeave) {
      this.props.children.props.onMouseLeave(ev);
    }
    this.props.onHover({ hovered: false, ...ev });
  };

  render() {
    return React.cloneElement(this.props.children, {
      onMouseEnter: this.onEnter,
      onMouseLeave: this.onLeave,
    });
  }
}

export default HoverController;
