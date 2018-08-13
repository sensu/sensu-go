import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import { RelativeToCurrentDate } from "/components/RelativeDate";

class EventStatusDescriptor extends React.PureComponent {
  static propTypes = {
    check: PropTypes.object,
    compact: PropTypes.bool,
    event: PropTypes.object.isRequired,
  };

  static defaultProps = {
    check: null,
    compact: false,
  };

  static fragments = {
    check: gql`
      fragment EventStatusDescriptor_check on Check {
        lastOK
        status
        occurrences
      }
    `,
    event: gql`
      fragment EventStatusDescriptor_event on Event {
        timestamp
      }
    `,
  };

  numFormatter = new Intl.NumberFormat();

  renderMetricDescription() {
    const { event } = this.props;

    return (
      <React.Fragment>
        Metric received
        <strong>
          <RelativeToCurrentDate dateTime={event.timestamp} />
        </strong>.
      </React.Fragment>
    );
  }

  renderOKDescription() {
    const { event } = this.props;

    return (
      <React.Fragment>
        Last executed{" "}
        <strong>
          <RelativeToCurrentDate dateTime={event.timestamp} />
        </strong>.
      </React.Fragment>
    );
  }

  renderIncidentDescription() {
    const { check, compact } = this.props;

    return (
      <React.Fragment>
        Incident started{" "}
        <strong>
          <RelativeToCurrentDate dateTime={check.lastOK} />
        </strong>
        {!compact && this.renderExtendedDescription()}
        .
      </React.Fragment>
    );
  }

  renderExtendedDescription() {
    const { check } = this.props;

    if (check.status > 2) {
      return (
        <React.Fragment>
          {" "}
          and the last execution exited with status{" "}
          <strong>{check.status}</strong>
        </React.Fragment>
      );
    }

    return (
      <React.Fragment>
        {" "}
        and has continued for{" "}
        <strong>{this.numFormatter.format(check.occurrences)}</strong>{" "}
        consecutive executions
      </React.Fragment>
    );
  }

  render() {
    const { check } = this.props;

    if (check === null) {
      return this.renderMetricDescription();
    }

    if (check.status === 0 || check.lastOK === null) {
      return this.renderOKDescription();
    }

    return this.renderIncidentDescription();
  }
}

export default EventStatusDescriptor;
