import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "material-ui/styles";
import Typography from "material-ui/Typography";
import OrganizationIcon from "../OrganizationIcon";

class EnvironmentLabelBase extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    organization: PropTypes.string.isRequired,
    environment: PropTypes.string.isRequired,
    icon: PropTypes.string.isRequired,
    iconColour: PropTypes.string.isRequired,
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
    const { classes, icon, iconColour, organization, environment } = this.props;

    return (
      <div className={classes.container}>
        <Typography className={classes.label} variant="subheading">
          <span className={classes.lighter}>{organization}</span>
          {" · "}
          <span className={classes.heavier}>{environment}</span>
        </Typography>
        <OrganizationIcon icon={icon} iconColour={iconColour} />
      </div>
    );
  }
}

const enhancer = withStyles(EnvironmentLabelBase.styles);
export default enhancer(EnvironmentLabelBase);
