import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { compose } from "lodash/fp";
import { withApollo } from "react-apollo";
import { withRouter } from "react-router-dom";
import Button from "@material-ui/core/Button";
import ConfirmDelete from "/components/partials/ConfirmDelete";
import deleteEntity from "/mutations/deleteEntity";

class EntityDetailsDeleteAction extends React.PureComponent {
  static propTypes = {
    client: PropTypes.object.isRequired,
    entity: PropTypes.object,
    history: PropTypes.object.isRequired,
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
        {dialog => (
          <Button variant="raised" onClick={() => dialog.open()}>
            Delete
          </Button>
        )}
      </ConfirmDelete>
    );
  }
}

const enhancer = compose(withRouter, withApollo);
export default enhancer(EntityDetailsDeleteAction);
