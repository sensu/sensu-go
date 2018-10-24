import * as React from "react";
import PropTypes from "prop-types";
import { withStyles } from "@material-ui/core/styles";
import { fade } from "@material-ui/core/styles/colorManipulator";

const styles = theme => ({
  root: {
    whiteSpace: "pre-wrap",
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
});

class Code extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    children: PropTypes.node,
  };

  static defaultProps = { children: null };

  render() {
    const { classes, children, ...props } = this.props;
    return (
      <code {...props} className={classes.root}>
        {children}
      </code>
    );
  }
}

export default withStyles(styles)(Code);
