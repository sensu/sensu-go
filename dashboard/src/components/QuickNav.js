import React from "react";
import PropTypes from "prop-types";
import classNames from "classnames";

import { withStyles } from "material-ui/styles";

import DashboardIcon from "material-ui-icons/Dashboard";
import EventIcon from "material-ui-icons/Notifications";
import EntityIcon from "material-ui-icons/DesktopMac";
import CheckIcon from "material-ui-icons/AssignmentTurnedIn";
import SilenceIcon from "material-ui-icons/VolumeOff";
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
      <div className={classNames(classes.quickNavContainer, className)}>
        <QuickNavButton Icon={DashboardIcon} caption="Dashboard" to="/" exact />
        <QuickNavButton Icon={EventIcon} caption="Events" to="events" />
        <QuickNavButton Icon={EntityIcon} caption="Entities" to="entities" />
        <QuickNavButton Icon={CheckIcon} caption="Checks" to="checks" />
        <QuickNavButton Icon={SilenceIcon} caption="Silences" to="silences" />
        <QuickNavButton Icon={HookIcon} caption="Hooks" to="hooks" />
        <QuickNavButton Icon={HandlerIcon} caption="Handlers" to="handlers" />
      </div>
    );
  }
}

export default withStyles(styles)(QuickNav);
