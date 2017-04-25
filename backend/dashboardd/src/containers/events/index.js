import React, {Component} from 'react';
import EventsList from 'components/eventsList'

export default class Events extends Component {
  constructor(props) {
    super(props);
    this.state = {events: []};
  }

  componentDidMount() {
    this.Events()
    .then(data => {
      this.setState({
        events : data
      });
    });
  }

  Events() {
    return fetch('/events', {
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
    return <EventsList events={this.state.events} />;
  }
}
