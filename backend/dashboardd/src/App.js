import React, { Component } from 'react';

class App extends Component {
  constructor(props) {
    super(props);
    this.state = {entities: []};
  }

  componentDidMount() {
    this.Entities()
    .then(data => {
      this.setState({
        entities : data
      });
    });
  }

  Entities() {
    return fetch('/entities', {
      headers : {
        'Content-Type': 'application/json',
        'Accept': 'application/json'
      }
    })
    .then(function(response){
      if (!response.ok) {
        throw Error(response.statusText);
      }
      return response.json();
    })
    .catch(function(error) {
      console.log(error);
    });
  }

  render() {
    return (
      <div>
      {
        this.state.entities.map(function(entity, i){
         return <li key={i}>{entity.id}</li>
        })
      }
      </div>
    );
  }
}

export default App;
