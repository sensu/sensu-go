import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Menu from "@material-ui/core/Menu";
import MenuItem from "@material-ui/core/MenuItem";

import ConfirmDelete from "/components/partials/ConfirmDelete";
import ListItem from "/components/partials/ListItem";

import RelativeDate from "/components/RelativeDate";
import CheckStatusIcon from "/components/CheckStatusIcon";
import NamespaceLink from "/components/util/NamespaceLink";

class EntitiesListItem extends React.PureComponent {
  static propTypes = {
    entity: PropTypes.object.isRequired,
    selected: PropTypes.bool,
    onChangeSelected: PropTypes.func,
    onClickDelete: PropTypes.func,
  };

  static defaultProps = {
    selected: undefined,
    onChangeSelected: ev => ev,
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
        system {
          platform
          platformVersion
        }
      }
    `,
  };

  renderMenu = ({ close, anchorEl }) => {
    const { onClickDelete } = this.props;

    return (
      <Menu open onClose={close} anchorEl={anchorEl}>
        <ConfirmDelete key="delete" onSubmit={onClickDelete}>
          {confirm => (
            <MenuItem
              onClick={() => {
                confirm.open();
                close();
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
    const { entity, selected, onChangeSelected } = this.props;

    return (
      <ListItem
        selected={selected}
        onChangeSelected={onChangeSelected}
        icon={<CheckStatusIcon statusCode={entity.status} />}
        title={
          <NamespaceLink
            namespace={entity.namespace}
            to={`/entities/${entity.name}`}
          >
            <strong>{entity.name}</strong> {entity.system.platform}{" "}
            {entity.system.platformVersion}
          </NamespaceLink>
        }
        details={
          <React.Fragment>
            <strong>{entity.class}</strong> - Last seen{" "}
            <strong>
              <RelativeDate dateTime={entity.lastSeen} />
            </strong>{" "}
            with status <strong>{entity.status}</strong>.
          </React.Fragment>
        }
        renderMenu={this.renderMenu}
      />
    );
  }
}

export default EntitiesListItem;
