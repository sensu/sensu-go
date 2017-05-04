import React from 'react';
import PropTypes from 'prop-types';
import MuiThemeProvider from 'material-ui/styles/MuiThemeProvider';
import getMuiTheme from 'material-ui/styles/getMuiTheme';

import Sidebar from 'components/sidebar';
import Toolbar from 'components/toolbar';

const muiTheme = getMuiTheme({
  palette: {
    primary1Color: '#92C72E',
  },
});

const styles = require('./app.css');

function App({ children }) {
  return (
    <MuiThemeProvider muiTheme={muiTheme}>
      <div>
        <Toolbar />
        <Sidebar />
        <div className={styles.content}>
          {children}
        </div>
      </div>
    </MuiThemeProvider>
  );
}

App.propTypes = {
  children: PropTypes.node.isRequired,
};

export default App;
