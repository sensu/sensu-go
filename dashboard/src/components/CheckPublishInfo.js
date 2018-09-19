import React from "react";
import PropTypes from "prop-types";

import CronDescriptor from "/components/partials/CronDescriptor";
import Code from "/components/Code";

class CheckPublishInfo extends React.Component {
  static propTypes = {
    check: PropTypes.object.isRequired,
  };

  render() {
    const { check } = this.props;

    return (
      <React.Fragment>
        <Code>{check.command}</Code>
        <br />
        Executed{" "}
        <strong>
          {check.interval ? (
            `
                        every
                        ${check.interval}
                        ${check.interval === 1 ? "second" : "seconds"}
                      `
          ) : (
            <CronDescriptor expression={check.cron} />
          )}
        </strong>{" "}
        by{" "}
        <strong>
          {check.subscriptions.length}{" "}
          {check.subscriptions.length === 1 ? "subscription" : "subscriptions"}
        </strong>.
      </React.Fragment>
    );
  }
}

export default CheckPublishInfo;
