import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Maybe from "/components/Maybe";
import { RelativeToCurrentDate } from "/components/RelativeDate";

class EntityStatusDescriptor extends React.PureComponent {
  static propTypes = {
    entity: PropTypes.object.isRequired,
  };

  static fragments = {
    entity: gql`
      fragment EntityStatusDescriptor_entity on Entity {
        lastSeen
        class
      }
    `,
  };

  render() {
    const { entity } = this.props;
    const lastSeen = (
      <Maybe value={entity.lastSeen} fallback="never">
        {val => <RelativeToCurrentDate dateTime={val} />}
      </Maybe>
    );

    return (
      <React.Fragment>
        The <strong>{entity.class}</strong> was last seen{" "}
        <strong>{lastSeen}</strong>.
      </React.Fragment>
    );
  }
}

export default EntityStatusDescriptor;
