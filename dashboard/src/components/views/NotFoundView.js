import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "@material-ui/core/styles";
import Typography from "@material-ui/core/Typography";
import Grid from "@material-ui/core/Grid";

const styles = theme => ({
  root: {
    flexGrow: 1,
    display: "flex",
    textAlign: "center",
    alignItems: "center",
    justifyContent: "center",
  },
  container: {
    maxWidth: 1080,
    margin: theme.spacing.unit,
  },
  headline: {
    fontSize: 100,
    fontWeight: 600,
    color: theme.palette.text.secondary,
  },
  body: {
    fontSize: 36,
    fontWeight: 300,
    color: theme.palette.text.hint,
    "& a": {
      textDecoration: "underline",
    },
    "& a:visited": {
      color: "inherit",
    },
    "& a:link": {
      color: "inherit",
    },
  },
  graphic: {
    margin: "-1em 0",
    fontSize: 80,
    color: theme.palette.text.hint,
    "& span": {
      color: theme.palette.text.primary,
    },
  },
  [theme.breakpoints.up("sm")]: {
    root: {
      textAlign: "left",
    },
    graphic: {
      display: "block",
      textAlign: "center",
      fontSize: 96,
      margin: 0,
    },
  },
});

class NotFoundView extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.objectOf(PropTypes.string).isRequired,
  };

  render() {
    const { classes } = this.props;

    return (
      <div className={classes.root}>
        <Grid container spacing={40} className={classes.container}>
          <Typography
            component={Grid}
            item
            xs={12}
            sm={6}
            className={classes.graphic}
            variant="headline"
          >
            <p>
              <span role="img" aria-label="ship">
                🚀
              </span>
              {"  ·  "}
              <span role="img" aria-label="moon">
                🌖
              </span>
            </p>
          </Typography>
          <Grid item xs={12} sm={6}>
            <Typography className={classes.headline} variant="headline">
              404
            </Typography>
            <Typography className={classes.body} variant="subheading">
              The page you requested was not found.{" "}
              <a href="#back" onClick={() => window.history.back()}>
                Go back
              </a>{" "}
              or <a href="/">return home</a>.
            </Typography>
          </Grid>
        </Grid>
      </div>
    );
  }
}

export default withStyles(styles)(NotFoundView);
