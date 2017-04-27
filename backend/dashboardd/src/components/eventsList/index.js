import React, {Component} from 'react';
import {Table, TableBody, TableHeader, TableHeaderColumn, TableRow, TableRowColumn} from 'material-ui/Table';

var styles = require('./eventsList.css');

export default class EventsList extends Component {
  constructor(props) {
    super(props);
    this.renderEvent= this.renderEvent.bind(this);
  }

  render() {
    return (
      <Table className={styles.table}>
        <TableHeader>
          <TableRow>
            <TableHeaderColumn>Entity</TableHeaderColumn>
            <TableHeaderColumn>Check</TableHeaderColumn>
            <TableHeaderColumn>Command</TableHeaderColumn>
            <TableHeaderColumn>Timestamp</TableHeaderColumn>
          </TableRow>
        </TableHeader>
        <TableBody>
          {Object.keys(this.props.events).map(this.renderEvent)}
        </TableBody>
      </Table>

    );
  }

  renderEvent(key) {
    return (
      <TableRow key={key}>
        <TableRowColumn>{this.props.events[key].entity.id}</TableRowColumn>
        <TableRowColumn>{this.props.events[key].check.name}</TableRowColumn>
        <TableRowColumn>{this.props.events[key].check.command}</TableRowColumn>
        <TableRowColumn>{this.props.events[key].timestamp}</TableRowColumn>
      </TableRow>
    );
  }
}
