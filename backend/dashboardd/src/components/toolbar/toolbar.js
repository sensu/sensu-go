import React, {Component} from 'react';
import AppBar from 'material-ui/AppBar';

var styles = require('./toolbar.css');

export default class Toolbar extends Component {
  render () {
    return (
      <div className={styles.wrapper}>
      <AppBar
        className={styles.appbar}
        showMenuIconButton={false}
        title="Dashboard"
        zDepth={0}
      />
      </div>
    )
  }
}
