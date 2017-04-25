import React from 'react';
import ReactDOM from 'react-dom';
import {
  BrowserRouter as Router,
  Route
} from 'react-router-dom'
import Events from 'containers/events';

ReactDOM.render(
  <Router>
    <Route path="/" component={Events}/>
  </Router>,
  document.getElementById('root')
);
