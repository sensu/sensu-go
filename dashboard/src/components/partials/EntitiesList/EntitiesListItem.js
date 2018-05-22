import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Typography from "@material-ui/core/Typography";

import RelativeDate from "/components/RelativeDate";
import StatusListItem from "/components/StatusListItem";
import NamespaceLink from "/components/util/NamespaceLink";

class EntitiesListItem extends React.PureComponent {
  static propTypes = {
    entity: PropTypes.object.isRequired,
    selected: PropTypes.bool,
    onClickSelect: PropTypes.func,
  };

  static defaultProps = {
    selected: undefined,
    onClickSelect: undefined,
  };

  static fragments = {
    entity: gql`
      fragment EntitiesListItem_entity on Entity {
        id
        name
        lastSeen
        class
        status
        system {
          platform
          platformVersion
        }
      }
    `,
  };

  render() {
    const { entity, selected, onClickSelect } = this.props;

    return (
      <StatusListItem
        status={entity.status}
        selected={selected}
        onClickSelect={onClickSelect}
        title={
          <NamespaceLink
            namespace={entity.namespace}
            to={`/entities/${entity.name}`}
          >
            <Typography color="textSecondary">
              <strong>{entity.name}</strong> {entity.system.platform}{" "}
              {entity.system.platformVersion}
            </Typography>
          </NamespaceLink>
        }
      >
        <strong>{entity.class}</strong> - Last seen{" "}
        <strong>
          <RelativeDate dateTime={entity.lastSeen} />
        </strong>{" "}
        with status <strong>{entity.status}</strong>.
      </StatusListItem>
    );
  }
}

export default EntitiesListItem;
