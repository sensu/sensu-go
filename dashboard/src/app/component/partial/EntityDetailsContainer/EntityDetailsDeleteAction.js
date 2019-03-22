import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { compose } from "recompose";
import { withApollo } from "react-apollo";
import { withRouter } from "react-router-dom";

import ConfirmDelete from "/app/component/partial/ConfirmDelete";
import deleteEntity from "/lib/mutation/deleteEntity";

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
        namespace
      }
    `,
  };

  deleteRecord = () => {
    const { client, entity } = this.props;
    const { id, namespace } = entity;

    // delete
    deleteEntity(client, { id });

    // optimistically redirect
    this.props.history.push(`/${namespace}/entities`);
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

const enhancer = compose(
  withRouter,
  withApollo,
);
export default enhancer(EntityDetailsDeleteAction);
