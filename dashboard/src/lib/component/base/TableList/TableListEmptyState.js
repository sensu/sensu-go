import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "@material-ui/core/styles";
import Typography from "@material-ui/core/Typography";

const styles = theme => ({
  root: {
    textAlign: "center",
    padding: theme.spacing.unit * 3,
    maxWidth: 480,
    margin: "0 auto",
    [theme.breakpoints.up("md")]: {
      paddingTop: theme.spacing.unit * 14,
      paddingBottom: theme.spacing.unit * 14,
    },
  },
  icon: {
    display: "inline-flex",
    justifyContent: "center",
    alignItems: "center",
    fontSize: "2rem",
  },
  line: {
    margin: theme.spacing.unit,
  },
});

export class TableListItem extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    primary: PropTypes.string.isRequired,
    secondary: PropTypes.string,
    loading: PropTypes.bool,
  };

  static defaultProps = {
    className: "",
    secondary: null,
    loading: false,
  };

  render() {
    const { classes, primary, secondary, loading } = this.props;

    return (
      <div
        className={classes.root}
        style={{ visibility: loading ? "hidden" : "visible" }}
      >
        <span
          className={`${classes.icon} ${classes.line}`}
          role="img"
          aria-label="exclaimation"
        >
          ⁉️
        </span>
        <Typography className={classes.line} variant="title">
          {primary}
        </Typography>
        <Typography className={classes.line} variant="body1">
          {secondary}
        </Typography>
      </div>
    );
  }
}

export default withStyles(styles)(TableListItem);
