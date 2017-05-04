import React from 'react';
import ReactDOM from 'react-dom';
import injectTapEventPlugin from 'react-tap-event-plugin';
import { BrowserRouter as Router, Route } from 'react-router-dom';

import App from 'containers/app';
import Events from 'containers/events';

import './index.css';

require('typeface-roboto');

injectTapEventPlugin();

ReactDOM.render(
  <Router>
    <App>
      <Route exact path="/" component={Events} />
    </App>
  </Router>,
  document.getElementById('root'),
);
