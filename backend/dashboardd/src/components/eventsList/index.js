import React, {Component} from 'react';

export default class EventsList extends Component {
  render() {
    return (
      <div>
      Events
      {
        this.props.events.map(this.renderEvent)
      }
      </div>
    );
  }

  renderEvent({event, i}) {
    return <li key={i}>{event}</li>;
  }
}
