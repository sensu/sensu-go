import React, { Component } from "react";
import logo from "./logo.png";
import "./App.css";

class App extends Component {
  render() {
    return (
      <div className="App">
        <header className="App-header">
          <img src={logo} className="App-logo" alt="logo" />
          <h1 className="App-title">Development!</h1>
        </header>
        <p className="App-intro">
          If you would like to run the web application run
          <code> yarn start</code> in <code>backend/dashboard</code>.
        </p>
      </div>
    );
  }
}

export default App;
