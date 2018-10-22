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
  rightMargin: {
    whiteSpace: "pre-wrap",
    marginRight: "24px",
  },
});

class Code extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    dictionaryMargin: PropTypes.bool,
  };

  static defaultProps = {
    dictionaryMargin: false,
  };

  render() {
    const { classes, dictionaryMargin, ...props } = this.props;
    const className = dictionaryMargin
      ? classes.root
      : classNames(classes.root, classes.rightMargin);
    return <code {...props} className={className} />;
  }
}

export default withStyles(styles)(Code);
