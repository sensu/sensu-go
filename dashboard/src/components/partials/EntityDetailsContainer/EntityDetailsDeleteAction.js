import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { compose } from "recompose";
import { withApollo } from "react-apollo";
import { withRouter } from "react-router-dom";

import ConfirmDelete from "/components/partials/ConfirmDelete";
import deleteEntity from "/mutations/deleteEntity";

class EntityDetailsDeleteAction extends React.PureComponent {
  static propTypes = {
    client: PropTypes.object.isRequired,
    entity: PropTypes.object,
    history: PropTypes.object.isRequired,
    children: PropTypes.func.isRequired,
  };

  static defaultProps = {
    entity: null,
  };

  static fragments = {
    entity: gql`
      fragment EntityDetailsDeleteAction_entity on Entity {
        id
        name
        ns: namespace {
          org: organization
          env: environment
        }
      }
    `,
  };

  deleteRecord = () => {
    // delete
    const { client, entity } = this.props;
    deleteEntity(client, entity);

    // optimistically redirect
    this.props.history.push(`/${entity.ns.org}/${entity.ns.env}/entities`);
  };

  render() {
    return (
      <ConfirmDelete
        identifier={this.props.entity.name}
        onSubmit={this.deleteRecord}
      >
        {dialog => this.props.children(dialog.open)}
      </ConfirmDelete>
    );
  }
}

const enhancer = compose(withRouter, withApollo);
export default enhancer(EntityDetailsDeleteAction);
