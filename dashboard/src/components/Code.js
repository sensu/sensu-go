import * as React from "react";
import PropTypes from "prop-types";
import { withStyles } from "@material-ui/core/styles";
import { fade } from "@material-ui/core/styles/colorManipulator";
import classNames from "classnames";

const styles = theme => ({
  root: {
    whiteSpace: "nowrap",
    margin: "0 1px",
    padding: `${2 / 16}em ${4 / 16}em`,
    borderRadius: `${5 / 16}em`,
    fontSize: `${14 / 16}em`,
    fontFamily: theme.typography.monospace.fontFamily,
    fontWeight: 500,
    userSelect: "text",
    backgroundColor: fade(theme.palette.text.primary, 0.05),
    color:
      theme.palette.type === "dark"
        ? theme.palette.secondary.light
        : theme.palette.secondary.dark,
  },
  preWrap: { whiteSpace: "pre-wrap" },
  block: {
    // because this seems to be layered with Monospaced
    // the background colours overlap otherwise
    backgroundColor: "unset",
  },
});

class Code extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    block: PropTypes.bool,
    preWrap: PropTypes.bool,
  };

  static defaultProps = {
    block: false,
    preWrap: false,
  };

  render() {
    const { classes, block, preWrap, ...props } = this.props;
    const className = classNames(classes.root, {
      [classes.block]: block,
      [classes.preWrap]: preWrap,
    });
    return <code {...props} className={className} />;
  }
}

export default withStyles(styles)(Code);
