import React from "react";
import PropTypes from "prop-types";

import SmallCheck from "/lib/component/icon/SmallCheck";
import Item from "@material-ui/core/MenuItem";
import ItemText from "@material-ui/core/ListItemText";
import ItemIcon from "@material-ui/core/ListItemIcon";

class Option extends React.PureComponent {
  static displayName = "ToolbarSelect.Option";

  static propTypes = {
    children: PropTypes.node,
    selected: PropTypes.bool,
    value: PropTypes.any.isRequired,
  };

  static defaultProps = {
    selected: false,
    children: null,
  };

  render() {
    const { value, selected, children, ...props } = this.props;
    const label = children || value;

    return (
      <Item key={value} value={value} {...props}>
        {selected && (
          <ItemIcon>
            <SmallCheck />
          </ItemIcon>
        )}
        <ItemText inset={selected} primary={label} />
      </Item>
    );
  }
}

export default Option;
