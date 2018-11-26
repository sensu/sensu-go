import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import capitalize from "lodash/capitalize";

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
        entityClass
      }
    `,
  };

  render() {
    const { entity } = this.props;

    if (!entity.lastSeen && entity.entityClass !== "agent") {
      return (
        <React.Fragment>
          <strong>{capitalize(entity.entityClass)}</strong> entity.
        </React.Fragment>
      );
    }

    return (
      <React.Fragment>
        The <strong>{entity.entityClass}</strong> was last seen{" "}
        <strong>
          <Maybe value={entity.lastSeen} fallback="never">
            {val => <RelativeToCurrentDate dateTime={val} />}
          </Maybe>
        </strong>.
      </React.Fragment>
    );
  }
}

export default EntityStatusDescriptor;
