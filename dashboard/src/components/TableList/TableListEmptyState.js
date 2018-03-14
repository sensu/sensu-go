import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "material-ui/styles";
import Typography from "material-ui/Typography";

const styles = theme => ({
  root: {
    textAlign: "center",
    paddingTop: theme.spacing.unit * 14,
    paddingBottom: theme.spacing.unit * 14,
    paddingLeft: theme.spacing.unit * 21,
    paddingRight: theme.spacing.unit * 21,
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
  };

  static defaultProps = {
    className: "",
    secondary: null,
  };

  render() {
    const { classes, primary, secondary } = this.props;

    return (
      <div className={classes.root}>
        <span
          className={`${classes.icon} ${classes.line}`}
          role="img"
          aria-label="exclaimation"
        >
          ⁉️
        </span>
        <Typography className={classes.line} type="title">
          {primary}
        </Typography>
        <Typography className={classes.line} type="body1">
          {secondary}
        </Typography>
      </div>
    );
  }
}

export default withStyles(styles)(TableListItem);
