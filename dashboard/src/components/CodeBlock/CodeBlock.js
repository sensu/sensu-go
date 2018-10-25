import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";
import { emphasize } from "@material-ui/core/styles/colorManipulator";
import Typography from "@material-ui/core/Typography";

const styles = theme => ({
  root: {
    fontFamily: theme.typography.monospace.fontFamily,
    overflowX: "scroll",
    userSelect: "text",
    tabSize: 2,
    color:
      theme.palette.type === "dark"
        ? theme.palette.secondary.light
        : theme.palette.secondary.dark,
  },
  background: {
    backgroundColor: emphasize(theme.palette.background.paper, 0.01875),
  },
  scaleFont: {
    // Browsers tend to render monospaced fonts a little larger than intended.
    // Attempt to scale accordingly.
    fontSize: "0.8125rem", // TODO: Scale given fontSize from theme?
  },
  wrap: { whiteSpace: "pre-wrap" },
});

class CodeBlock extends React.Component {
  static propTypes = {
    background: PropTypes.bool,
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    component: PropTypes.oneOfType([PropTypes.string, PropTypes.func]),
    children: PropTypes.node.isRequired,
    scaleFont: PropTypes.bool,
  };

  static defaultProps = {
    background: true,
    component: "pre",
    className: "",
    scaleFont: true,
    highlight: null,
  };

  render() {
    const {
      background,
      classes,
      className: classNameProp,
      children,
      scaleFont,
      ...props
    } = this.props;

    const className = classnames(classes.root, classNameProp, {
      [classes.background]: background,
      [classes.scaleFont]: scaleFont,
    });

    return (
      <Typography className={className} {...props}>
        <code className={classes.wrap}>{children}</code>
      </Typography>
    );
  }
}

export default withStyles(styles)(CodeBlock);
