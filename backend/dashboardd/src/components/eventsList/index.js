import React, {Component} from 'react';

export default class EventsList extends Component {
  constructor(props) {
    super(props);
    this.renderEvent= this.renderEvent.bind(this);
  }

  render() {
    return (
      <div>
      Events
      <table class="pure-table">
        <thead>
          <tr>
            <th>Entity</th>
            <th>Check</th>
            <th>Command</th>
            <th>Timestamp</th>
          </tr>
        </thead>

        <tbody>
          {Object.keys(this.props.events).map(this.renderEvent)}
        </tbody>
      </table>

      </div>
    );
  }

  renderEvent(key) {
    return (
      <tr key={key}>
        <td>{this.props.events[key].entity.id}</td>
        <td>{this.props.events[key].check.name}</td>
        <td>{this.props.events[key].check.command}</td>
        <td>{this.props.events[key].timestamp}</td>
      </tr>
    );
  }
}
