import React, {Component} from 'react';
import Drawer from 'material-ui/Drawer';
import MenuItem from 'material-ui/MenuItem';
import {Link} from 'react-router-dom'

var styles = require('./sidebar.css');

export default class Sidebar extends Component {
  render () {
    return (
      <div className={styles.sidebar}>
        <Drawer
          docked={true}
          open={true}
          zDepth={0}
        >
          <div className={styles.logo}>
            <img alt="sensu" src="logo.png"/>
          </div>
          <nav>
            <MenuItem
              containerElement={<Link to="/" />}
              primaryText="Events"
            />
          </nav>
        </Drawer>
      </div>
    )
  }
}
