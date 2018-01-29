import React from "react";
import PropTypes from "prop-types";
import classNames from "classnames";

import { withStyles } from "material-ui/styles";

import DashboardIcon from "material-ui-icons/Dashboard";
import EventIcon from "material-ui-icons/Notifications";
import EntityIcon from "material-ui-icons/DesktopMac";
import CheckIcon from "material-ui-icons/AssignmentTurnedIn";
import SilencedIcon from "material-ui-icons/VolumeOff";
import HookIcon from "material-ui-icons/Link";
import HandlerIcon from "material-ui-icons/CallSplit";

import QuickNavButton from "./QuickNavButton";

const styles = {
  quicknavcontainer: { padding: 0 },
};

class QuickNav extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
  };

  static defaultProps = { className: "" };

  render() {
    const { classes, className } = this.props;
    return (
      <div className={classNames(classes.quicknavcontainer, className)}>
        <QuickNavButton
          className={classes.quicknavbutton}
          Icon={DashboardIcon}
          primary="Dashboard"
        />
        <QuickNavButton
          className={classes.quicknavbutton}
          Icon={EventIcon}
          primary="Events"
          href="/events"
        />
        <QuickNavButton
          className={classes.quicknavbutton}
          Icon={EntityIcon}
          primary="Entities"
        />
        <QuickNavButton
          className={classes.quicknavbutton}
          Icon={CheckIcon}
          primary="Checks"
          href="/checks"
        />
        <QuickNavButton
          className={classes.quicknavbutton}
          Icon={SilencedIcon}
          primary="Silenced"
          href="/silenced"
        />
        <QuickNavButton
          className={classes.quicknavbutton}
          Icon={HookIcon}
          primary="Hooks"
          href="/hooks"
        />
        <QuickNavButton
          className={classes.quicknavbutton}
          Icon={HandlerIcon}
          primary="Handlers"
          href="/handlers"
        />
      </div>
    );
  }
}

export default withStyles(styles)(QuickNav);
