import React, {Component} from 'react';
import {Table, TableBody, TableHeader, TableHeaderColumn, TableRow} from 'material-ui/Table';
import map from 'lodash/map';

import EventRow from 'components/eventRow'

var styles = require('./eventsList.css');

export default class EventsList extends Component {
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
          {map(this.props.events, (event, i) => (
            <EventRow key={i} {...event} />
          ))}
        </TableBody>
      </Table>

    );
  }
}
