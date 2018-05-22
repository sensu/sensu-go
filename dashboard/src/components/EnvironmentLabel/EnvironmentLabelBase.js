import React from "react";
import PropTypes from "prop-types";
import Typography from "@material-ui/core/Typography";
import { withStyles } from "@material-ui/core/styles";

import { EnvironmentIconBase as Icon } from "/components/EnvironmentIcon";

class EnvironmentLabelBase extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    organization: PropTypes.string.isRequired,
    organizationIcon: PropTypes.string.isRequired,
    environment: PropTypes.string.isRequired,
    environmentColour: PropTypes.string.isRequired,
  };

  static styles = theme => ({
    container: {
      display: "flex",
      alignSelf: "center",
    },
    label: {
      color: "inherit",
      marginRight: theme.spacing.unit,
      opacity: 0.88,
    },
    heavier: {
      fontWeight: 400,
    },
    lighter: {
      fontWeight: 300,
      opacity: 0.71,
    },
  });

  render() {
    const {
      classes,
      organizationIcon,
      environmentColour,
      organization,
      environment,
    } = this.props;

    return (
      <div className={classes.container}>
        <Typography className={classes.label} variant="subheading">
          <span className={classes.lighter}>{organization}</span>
          {" · "}
          <span className={classes.heavier}>{environment}</span>
        </Typography>

        <Icon
          organizationIcon={organizationIcon}
          environmentColour={environmentColour}
        />
      </div>
    );
  }
}

const enhancer = withStyles(EnvironmentLabelBase.styles);
export default enhancer(EnvironmentLabelBase);
