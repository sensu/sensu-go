import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { compose } from "recompose";
import { withApollo } from "react-apollo";
import { withRouter } from "react-router-dom";

import ConfirmDelete from "/app/component/partial/ConfirmDelete";
import deleteCheck from "/lib/mutation/deleteCheck";

class CheckDetailsDeleteAction extends React.PureComponent {
  static propTypes = {
    check: PropTypes.object,
    children: PropTypes.func.isRequired,
    client: PropTypes.object.isRequired,
    history: PropTypes.object.isRequired,
  };

  static defaultProps = {
    check: null,
  };

  static fragments = {
    check: gql`
      fragment CheckDetailsDeleteAction_check on CheckConfig {
        id
        name
        namespace
      }
    `,
  };

  deleteRecord = () => {
    const { client, check, history } = this.props;
    const { id, namespace } = check;

    // Send request
    deleteCheck(client, { id });

    // Optimistically redirect
    history.replace(`/${namespace}/checks`);
  };

  render() {
    return (
      <ConfirmDelete
        identifier={this.props.check.name}
        onSubmit={this.deleteRecord}
      >
        {dialog => this.props.children(dialog.open)}
      </ConfirmDelete>
    );
  }
}

const enhancer = compose(
  withApollo,
  withRouter,
);
export default enhancer(CheckDetailsDeleteAction);
