import React from "react";
import PropTypes from "prop-types";

import Drawer from "material-ui/Drawer";
import List, { ListItem, ListItemText } from "material-ui/List";
import Avatar from "material-ui/Avatar";

import LinkIcon from "material-ui-icons/Link";
import WorkIcon from "material-ui-icons/Work";
import ZoomIcon from "material-ui-icons/ZoomOut";

const logo = require("../assets/logo.png");

function Sidebar({ open }) {
  return (
    <Drawer type="persistent" open={open}>
      <div>
        <img
          alt="sensu"
          src={logo}
          style={{ width: "200px", padding: "1em" }}
        />
      </div>
      <nav>
        <List>
          <ListItem button>
            <Avatar>
              <LinkIcon />
            </Avatar>
            <ListItemText primary="Events" />
          </ListItem>
          <ListItem button>
            <Avatar>
              <WorkIcon />
            </Avatar>
            <ListItemText primary="Entities" />
          </ListItem>
          <ListItem button>
            <Avatar>
              <ZoomIcon />
            </Avatar>
            <ListItemText primary="Checks" />
          </ListItem>
        </List>
      </nav>
    </Drawer>
  );
}

Sidebar.propTypes = {
  open: PropTypes.bool.isRequired,
};

export default Sidebar;
