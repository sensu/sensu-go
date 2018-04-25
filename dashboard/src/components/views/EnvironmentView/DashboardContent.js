import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "material-ui/styles";
import Grid from "material-ui/Grid";

import AppContent from "/components/AppContent";
import Placeholder from "/components/PlaceholderCard";

const styles = () => ({
  root: {
    flexGrow: 1,
  },
});

class DashboardContent extends React.Component {
  static propTypes = {
    classes: PropTypes.objectOf(PropTypes.string).isRequired,
  };

  render() {
    const { classes } = this.props;
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
