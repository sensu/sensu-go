import React from 'react';
import Drawer from 'material-ui/Drawer';
import MenuItem from 'material-ui/MenuItem';
import { Link } from 'react-router-dom';

const styles = require('./sidebar.css');

function Sidebar() {
  return (
    <div className={styles.sidebar}>
      <Drawer
        docked
        open
        zDepth={0}
      >
        <div className={styles.logo}>
          <img alt="sensu" src="logo.png" />
        </div>
        <nav>
          <MenuItem
            containerElement={<Link to="/" />}
            primaryText="Events"
          />
        </nav>
      </Drawer>
    </div>
  );
}

export default Sidebar;
