import React from "react";
import PropTypes from "prop-types";
import Typography from "@material-ui/core/Typography";
import { withStyles } from "@material-ui/core/styles";
import { emphasize } from "@material-ui/core/styles/colorManipulator";

const styles = theme => ({
  root: {
    borderRadius: theme.spacing.unit / 2,
    padding: `${theme.spacing.unit / 8}px ${theme.spacing.unit / 2}px`,
    background: emphasize(theme.palette.background.paper, 0.05),
    color: emphasize(theme.palette.text.primary, 0.22),
    display: "inline",
    whiteSpace: "nowrap",
  },
  value: {
    color: theme.palette.text.primary,
  },
});

class Label extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    name: PropTypes.string.isRequired,
    value: PropTypes.string.isRequired,
  };

  render() {
    const { classes, name, value } = this.props;
    return (
      <Typography component="span" className={classes.root}>
        {name} | <span className={classes.value}>{value}</span>
      </Typography>
    );
  }
}

export default withStyles(styles)(Label);
