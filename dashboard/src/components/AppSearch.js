import React from "react";
import PropTypes from "prop-types";
import compose from "recompose/compose";
import pure from "recompose/pure";
import withWidth, { isWidthUp } from "material-ui/utils/withWidth";
import SearchIcon from "material-ui-icons/Search";
import { fade } from "material-ui/styles/colorManipulator";
import { withStyles } from "material-ui/styles";

const styles = theme => ({
  wrapper: {
    fontFamily: theme.typography.fontFamily,
    position: "relative",
    borderRadius: 2,
    background: fade(theme.palette.common.white, 0.15),
    "&:hover": {
      background: fade(theme.palette.common.white, 0.25),
    },
    "& $input": {
      transition: theme.transitions.create("width"),
      width: 200,
      "&:focus": {
        width: 250,
      },
    },
  },
  search: {
    width: theme.spacing.unit * 6,
    height: "100%",
    position: "absolute",
    pointerEvents: "none",
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
  },
  input: {
    font: "inherit",
    padding: `
      ${theme.spacing.unit}px ${theme.spacing.unit}px
      ${theme.spacing.unit}px ${theme.spacing.unit * 6}px
    `,
    border: 0,
    display: "block",
    verticalAlign: "middle",
    whiteSpace: "normal",
    background: "none",
    margin: 0, // Reset for Safari
    color: "inherit",
    width: "100%",
    "&:focus": {
      outline: 0,
    },
    "&::placeholder": {
      color: fade(theme.palette.common.white, 0.4),
    },
  },
});

class DrawerSearch extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    width: PropTypes.string.isRequired,
    className: PropTypes.string,
  };

  static defaultProps = {
    className: "",
  };

  render() {
    const { className, classes, width } = this.props;

    if (!isWidthUp("sm", width)) {
      return null;
    }

    return (
      <div className={`${classes.wrapper} ${className}`}>
        <div className={classes.search}>
          <SearchIcon />
        </div>
        <input className={classes.input} placeholder="Search" />
      </div>
    );
  }
}

export default compose(
  withStyles(styles, { name: "DrawerSearch" }),
  withWidth(),
  pure,
)(DrawerSearch);
