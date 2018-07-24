import React from "react";
import PropTypes from "prop-types";

import Menu from "@material-ui/core/Menu";
import MenuItem from "@material-ui/core/MenuItem";
import ListItemIcon from "@material-ui/core/ListItemIcon";
import ListItemText from "@material-ui/core/ListItemText";

import CheckStatusIcon from "/components/CheckStatusIcon";

class StatusMenu extends React.Component {
  static propTypes = {
    anchorEl: PropTypes.string,
    className: PropTypes.string,
    onChange: PropTypes.func,
    onClose: PropTypes.func,
  };

  static defaultProps = {
    anchorEl: undefined,
    className: undefined,
    onChange: undefined,
    onClose: undefined,
  };

  render() {
    const { anchorEl, className, onClose, onChange } = this.props;

    return (
      <Menu anchorEl={anchorEl} className={className} onClose={onClose} open>
        <MenuItem
          key="incident"
          onClick={() => onChange("HasCheck && IsIncident")}
        >
          <ListItemText primary="Incident" style={{ paddingLeft: 40 }} />
        </MenuItem>
        <MenuItem key="warning" onClick={() => onChange([1])}>
          <ListItemIcon>
            <CheckStatusIcon statusCode={1} />
          </ListItemIcon>
          <ListItemText primary="Warning" />
        </MenuItem>
        <MenuItem key="critical" onClick={() => onChange([2])}>
          <ListItemIcon>
            <CheckStatusIcon statusCode={2} />
          </ListItemIcon>
          <ListItemText primary="Critical" />
        </MenuItem>
        <MenuItem key="unknown" onClick={() => onChange([3])}>
          <ListItemIcon>
            <CheckStatusIcon statusCode={3} />
          </ListItemIcon>
          <ListItemText primary="Unknown" />
        </MenuItem>
        <MenuItem key="passing" onClick={() => onChange([0])}>
          <ListItemIcon>
            <CheckStatusIcon statusCode={0} />
          </ListItemIcon>
          <ListItemText primary="Passing" />
        </MenuItem>
      </Menu>
    );
  }
}

export default StatusMenu;
