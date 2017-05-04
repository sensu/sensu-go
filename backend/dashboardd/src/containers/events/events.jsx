import React, { Component } from 'react';
import EventsList from 'components/eventsList';

export default class Events extends Component {
  static fetch() {
    return fetch('/events', {
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
      },
    })
    .then((response) => {
      if (!response.ok) {
        throw Error(response.statusText);
      }
      return response.json();
    });
  }

  constructor(props) {
    super(props);
    this.state = { events: [] };
  }

  componentDidMount() {
    Events.fetch()
    .then((data) => {
      this.setState({
        events: data,
      });
    });
  }

  render() {
    return <EventsList events={this.state.events} />;
  }
}
