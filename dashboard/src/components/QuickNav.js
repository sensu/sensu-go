import React from "react";
// import PropTypes from "prop-types";

import { styles as listItemIconStyles } from "material-ui/List/ListItemIcon";
import { withStyles } from "material-ui/styles";

import DashboardIcon from "material-ui-icons/Dashboard";
import EventIcon from "material-ui-icons/Notifications";
import EntityIcon from "material-ui-icons/DesktopMac";
import CheckIcon from "material-ui-icons/AssignmentTurnedIn";
import SilencedIcon from "material-ui-icons/VolumeOff";
import HookIcon from "material-ui-icons/Link";
import HandlerIcon from "material-ui-icons/CallSplit";

import QuickNavButton from "./QuickNavButton";

const styles = theme => {
  const listItemStyles = listItemIconStyles(theme);

  return {
    listItemButton: listItemStyles.root, // TODO ...
    listItemContent: {
      flex: "1 1 auto",
      padding: "0 16px",
      height: listItemStyles.root.height,
      "&:first-child": {
        paddingLeft: theme.spacing.unit * 7,
      },
    },
  };
};

class QuickNav extends React.Component {
  render() {
    return (
      <div className="QuickNav" style={{ padding: "90px 0" }}>
        <QuickNavButton Icon={DashboardIcon} primary="Dashboard" />
        <QuickNavButton Icon={EventIcon} primary="Events" href="/events" />
        <QuickNavButton Icon={EntityIcon} primary="Entities" />
        <QuickNavButton Icon={CheckIcon} primary="Checks" href="/checks" />
        <QuickNavButton
          Icon={SilencedIcon}
          primary="Silenced"
          href="/silenced"
        />
        <QuickNavButton Icon={HookIcon} primary="Hooks" href="/hooks" />
        <QuickNavButton
          Icon={HandlerIcon}
          primary="Handlers"
          href="/handlers"
        />
      </div>
    );
  }
}

export default withStyles(styles)(QuickNav);
