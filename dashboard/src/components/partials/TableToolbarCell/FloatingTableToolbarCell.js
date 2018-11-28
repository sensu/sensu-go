import React from "react";
import PropTypes from "prop-types";

import Media from "react-media";
import TableToolbarCell from "./TableToolbarCell";

class FloatingTableToolbarCell extends React.Component {
  static propTypes = {
    disabled: PropTypes.bool,
    hovered: PropTypes.bool,
    children: PropTypes.func.isRequired,
  };

  static defaultProps = {
    disabled: false,
    hovered: false,
  };

  renderCell = ({ canHover, ...props }) => {
    const { hovered, disabled: disabledProp, children } = props;
    const disabled = disabledProp || (canHover && !hovered);

    return (
      <TableToolbarCell floating={canHover} disabled={disabled}>
        {children}
      </TableToolbarCell>
    );
  };

  render() {
    return (
      <Media query="screen and (any-hover)">
        {canHover => this.renderCell({ canHover, ...this.props })}
      </Media>
    );
  }
}

export default FloatingTableToolbarCell;
