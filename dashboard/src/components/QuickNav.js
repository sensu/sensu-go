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
  quickNavContainer: {},
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
      <div className={classNames(classes.quickNavcContainer, className)}>
        <QuickNavButton
          Icon={DashboardIcon}
          primary="Dashboard"
          href="/"
          active={location.pathname === "/"}
        />
        <QuickNavButton
          Icon={EventIcon}
          primary="Events"
          href="/events"
          active={location.pathname === "/events"}
        />
        <QuickNavButton
          Icon={EntityIcon}
          primary="Entities"
          href="/entities"
          active={location.pathname === "/entities"}
        />
        <QuickNavButton
          Icon={CheckIcon}
          primary="Checks"
          href="/checks"
          active={location.pathname === "/checks"}
        />
        <QuickNavButton
          Icon={SilencedIcon}
          primary="Silences"
          href="/silences"
          active={location.pathname === "/silences"}
        />
        <QuickNavButton
          Icon={HookIcon}
          primary="Hooks"
          href="/hooks"
          active={location.pathname === "/hooks"}
        />
        <QuickNavButton
          Icon={HandlerIcon}
          primary="Handlers"
          href="/handlers"
          active={location.pathname === "/handlers"}
        />
      </div>
    );
  }
}

export default withStyles(styles)(QuickNav);
