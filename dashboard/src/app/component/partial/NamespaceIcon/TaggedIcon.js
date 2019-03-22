import React from "react";
import PropTypes from "prop-types";

import Tag from "/lib/component/base/TagOrb";
import Icon from "./Icon";

class TaggedIcon extends React.Component {
  static displayName = "NamespaceIcon.TaggedIcon";

  static propTypes = {
    icon: PropTypes.string.isRequired,
    colour: PropTypes.string.isRequired,
    size: PropTypes.number,
  };

  static defaultProps = {
    size: 24.0,
  };

  render() {
    const { size, icon, colour, ...props } = this.props;

    return (
      <Icon icon={icon} size={size} {...props}>
        <Tag colour={colour} size={size / 3.0} />
      </Icon>
    );
  }
}

export default TaggedIcon;
