import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "material-ui/styles";
import { emphasize } from "material-ui/styles/colorManipulator";
import Typography from "material-ui/Typography";

const styles = theme => ({
  root: {
    // TODO: Move into theme so that it can be overridden.
    fontFamily: `"SFMono-Regular",Consolas,"Liberation Mono",Menlo,Courier,monospace`,
    overflowX: "scroll",
    userSelect: "text",
    tabSize: 2,
  },
  background: {
    backgroundColor: emphasize(theme.palette.background.paper, 0.01875),
  },
  scaleFont: {
    // Browsers tend to render monospaced fonts a little larger than intended.
    // Attempt to scale accordingly.
    fontSize: "0.8125rem", // TODO: Scale given fontSize from theme?
  },
});

class Monospaced extends React.Component {
  static propTypes = {
    background: PropTypes.bool,
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    component: PropTypes.oneOfType([PropTypes.string, PropTypes.func]),
    children: PropTypes.node.isRequired,
    scaleFont: PropTypes.bool,
  };

  static defaultProps = {
    background: false,
    component: "pre",
    className: "",
    scaleFont: true,
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
        {children}
      </Typography>
    );
  }
}

export default withStyles(styles)(Monospaced);
