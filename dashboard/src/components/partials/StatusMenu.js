import React from "react";
import PropTypes from "prop-types";

import MenuItem from "@material-ui/core/MenuItem";
import ListItemIcon from "@material-ui/core/ListItemIcon";
import ListItemText from "@material-ui/core/ListItemText";

import ButtonMenu from "/components/partials/ButtonMenu";
import CheckStatusIcon from "/components/CheckStatusIcon";
import ErrorHollow from "/icons/ErrorHollow";

class StatusMenu extends React.Component {
  static propTypes = {
    className: PropTypes.string,
    onChange: PropTypes.func,
  };

  static defaultProps = {
    className: undefined,
    onChange: undefined,
  };

  render() {
    const { className, onChange } = this.props;

    return (
      <ButtonMenu className={className} label="Status" onChange={onChange}>
        <MenuItem key="incident" value={"HasCheck && IsIncident"}>
          <ListItemIcon>
            <ErrorHollow />
          </ListItemIcon>
          <ListItemText primary="Incident" />
        </MenuItem>
        <MenuItem key="warning" value={[1]}>
          <ListItemIcon>
            <CheckStatusIcon statusCode={1} />
          </ListItemIcon>
          <ListItemText primary="Warning" />
        </MenuItem>
        <MenuItem key="critical" value={[2]}>
          <ListItemIcon>
            <CheckStatusIcon statusCode={2} />
          </ListItemIcon>
          <ListItemText primary="Critical" />
        </MenuItem>
        <MenuItem key="unknown" value={[3]}>
          <ListItemIcon>
            <CheckStatusIcon statusCode={3} />
          </ListItemIcon>
          <ListItemText primary="Unknown" />
        </MenuItem>
        <MenuItem key="passing" value={[0]}>
          <ListItemIcon>
            <CheckStatusIcon statusCode={0} />
          </ListItemIcon>
          <ListItemText primary="Passing" />
        </MenuItem>
      </ButtonMenu>
    );
  }
}

export default StatusMenu;
