import React from 'react';
import AppBar from 'material-ui/AppBar';

const styles = require('./toolbar.css');

function Toolbar() {
  return (
    <div className={styles.wrapper}>
      <AppBar
        className={styles.appbar}
        showMenuIconButton={false}
        title="Dashboard"
        zDepth={0}
      />
    </div>
  );
}

export default Toolbar;
