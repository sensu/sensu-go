import React from "react";
import PropTypes from "prop-types";
import classNames from "classnames";

import { withStyles } from "@material-ui/core/styles";
import EventIcon from "@material-ui/icons/Notifications";
import EntityIcon from "@material-ui/icons/DesktopMac";
import CheckIcon from "@material-ui/icons/AssignmentTurnedIn";

import QuickNavButton from "/components/QuickNavButton";

const styles = {
  quickNavContainer: {},
};

class QuickNav extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    organization: PropTypes.string.isRequired,
    environment: PropTypes.string.isRequired,
  };

  static defaultProps = { className: "" };

  render() {
    const { classes, className, organization, environment } = this.props;

    return (
      <div className={classNames(classes.quickNavContainer, className)}>
        <QuickNavButton
          organization={organization}
          environment={environment}
          Icon={EventIcon}
          caption="Events"
          to="events"
        />
        <QuickNavButton
          organization={organization}
          environment={environment}
          Icon={EntityIcon}
          caption="Entities"
          to="entities"
        />
        <QuickNavButton
          organization={organization}
          environment={environment}
          Icon={CheckIcon}
          caption="Checks"
          to="checks"
        />
      </div>
    );
  }
}

export default withStyles(styles)(QuickNav);
