import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "@material-ui/core/styles";
import { fade } from "@material-ui/core/styles/colorManipulator";

import KebabIcon from "/icons/Kebab";
import IconButton from "@material-ui/core/IconButton";

const styles = theme => ({
  root: { color: theme.palette.text.secondary },
  active: {
    backgroundColor: fade(
      theme.palette.action.active,
      theme.palette.action.hoverOpacity,
    ),
  },
});

class OverflowButton extends React.PureComponent {
  static propTypes = {
    active: PropTypes.bool,
    classes: PropTypes.object.isRequired,
    idx: PropTypes.string.isRequired,
    onClick: PropTypes.func.isRequired,
  };

  static defaultProps = {
    active: true,
  };

  render() {
    const { active, classes, idx, onClick } = this.props;

    return (
      <IconButton
        aria-label="More options"
        aria-owns={idx}
        aria-haspopup="true"
        onClick={onClick}
        className={active ? classes.active : ""}
      >
        <KebabIcon />
      </IconButton>
    );
  }
}

export default withStyles(styles)(OverflowButton);
