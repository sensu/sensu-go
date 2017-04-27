import React, {Component} from 'react';
import MuiThemeProvider from 'material-ui/styles/MuiThemeProvider';
import getMuiTheme from 'material-ui/styles/getMuiTheme';

import Sidebar from 'components/sidebar'
import Toolbar from 'components/toolbar'

const muiTheme = getMuiTheme({
  palette: {
    primary1Color: '#92C72E'
  }
});

var styles = require('./app.css');

export default class App extends Component {
  render() {
    return (
      <MuiThemeProvider muiTheme={muiTheme}>
        <div>
          <Toolbar/>
          <Sidebar/>
          <div className={styles.content}>
            {this.props.children}
          </div>
        </div>
      </MuiThemeProvider>
    );
  }
}
