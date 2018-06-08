import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Menu from "@material-ui/core/Menu";
import MenuItem from "@material-ui/core/MenuItem";
import Typography from "@material-ui/core/Typography";

import ConfirmDelete from "/components/partials/ConfirmDelete";
import RelativeDate from "/components/RelativeDate";
import StatusListItem from "/components/StatusListItem";
import NamespaceLink from "/components/util/NamespaceLink";

class EntitiesListItem extends React.PureComponent {
  static propTypes = {
    entity: PropTypes.object.isRequired,
    selected: PropTypes.bool,
    onClickSelect: PropTypes.func,
    onClickDelete: PropTypes.func,
  };

  static defaultProps = {
    selected: undefined,
    onClickSelect: ev => ev,
    onClickDelete: ev => ev,
  };

  static fragments = {
    entity: gql`
      fragment EntitiesListItem_entity on Entity {
        id
        name
        lastSeen
        class
        status
        deleted @client
        system {
          platform
          platformVersion
        }
      }
    `,
  };

  _renderMenu = renderProps => {
    const { open, onClose, anchorEl } = renderProps;
    const { onClickDelete } = this.props;

    return (
      <Menu keepMounted open={open} onClose={onClose} anchorEl={anchorEl}>
        <ConfirmDelete key="delete" onSubmit={onClickDelete}>
          {confirm => (
            <MenuItem
              onClick={() => {
                confirm.open();
                onClose();
              }}
            >
              Delete
            </MenuItem>
          )}
        </ConfirmDelete>
      </Menu>
    );
  };

  render() {
    const { entity, selected, onClickSelect } = this.props;

    // NOTE: Replace this when we add pagination to lists.
    if (entity.deleted) {
      return null;
    }

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
        renderMenu={this._renderMenu}
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
