import React from "react";
import PropTypes from "prop-types";

import InputAdornment from "@material-ui/core/InputAdornment";
import IconButton from "@material-ui/core/IconButton";
import Undo from "@material-ui/icons/Undo";

class ResetAdornment extends React.PureComponent {
  static propTypes = {
    onClick: PropTypes.func,
  };

  static defaultProps = {
    onClick: undefined,
  };

  render() {
    const { onClick } = this.props;

    return (
      <InputAdornment position="end">
        <IconButton
          aria-label="reset field to initial value"
          onClick={onClick}
          onMouseDown={event => event.preventDefault()}
        >
          <Undo />
        </IconButton>
      </InputAdornment>
    );
  }
}

export default ResetAdornment;
