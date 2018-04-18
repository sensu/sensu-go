import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "material-ui/styles";

const styles = theme => ({
  root: {
    height: 2,
    flexShrink: 0,
    border: "none",
    margin: "0 0 -2px 0",
  },
  green: {
    backgroundColor: theme.palette.green,
  },
  yellow: {
    backgroundColor: theme.palette.yellow,
  },
  orange: {
    backgroundColor: theme.palette.orange,
  },
  red: {
    backgroundColor: theme.palette.red,
  },
});

class CardHighlight extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    color: PropTypes.oneOf(["green", "yellow", "orange", "red"]),
  };

  static defaultProps = {
    color: "green",
  };

  render() {
    const { classes, color } = this.props;
    const className = classnames(classes.root, classes[color]);

    return <hr className={className} />;
  }
}

export default withStyles(styles)(CardHighlight);
