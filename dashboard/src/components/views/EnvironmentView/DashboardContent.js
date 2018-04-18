import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "material-ui/styles";
import Typography from "material-ui/Typography";
import Grid from "material-ui/Grid";
import Paper from "material-ui/Paper";

import AppContent from "../../AppContent";

const styles = () => ({
  root: {
    flexGrow: 1,
  },
  card: {
    display: "flex",
    minHeight: 180,
    alignItems: "center",
    justifyContent: "center",
  },
  tall: {
    minHeight: 360,
  },
});

class DashboardContent extends React.Component {
  static propTypes = {
    classes: PropTypes.objectOf(PropTypes.string).isRequired,
  };

  render() {
    const { classes } = this.props;
    const Placeholder = ({ tall, ...props }) => (
      <Typography
        component={Paper}
        variant="body1"
        className={classnames(classes.card, { [classes.tall]: tall })}
        {...props}
      >
        [ placeholder ]
      </Typography>
    );

    return (
      <AppContent className={classes.root} fullWidth gutters>
        <Grid container spacing={8}>
          <Grid item xs={12}>
            <Placeholder tall />
          </Grid>
          <Grid item xs={12} md={6}>
            <Placeholder />
          </Grid>
          <Grid item xs={12} md={6}>
            <Placeholder />
          </Grid>
          <Grid item xs={12} md={3}>
            <Placeholder />
          </Grid>
          <Grid item xs={12} md={3}>
            <Placeholder />
          </Grid>
          <Grid item xs={12} md={3}>
            <Placeholder />
          </Grid>
          <Grid item xs={12} md={3}>
            <Placeholder />
          </Grid>
        </Grid>
      </AppContent>
    );
  }
}

export default withStyles(styles)(DashboardContent);
