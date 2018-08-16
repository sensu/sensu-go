import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { compose } from "recompose";
import { withApollo } from "react-apollo";
import { withRouter } from "react-router-dom";

import ConfirmDelete from "/components/partials/ConfirmDelete";
import deleteCheck from "/mutations/deleteCheck";

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
        ns: namespace {
          org: organization
          env: environment
        }
      }
    `,
  };

  deleteRecord = () => {
    const { client, check, history } = this.props;
    const { id, ns } = check;

    // Send request
    deleteCheck(client, { id });

    // Optimistically redirect
    history.replace(`/${ns.org}/${ns.env}/checks`);
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

const enhancer = compose(withApollo, withRouter);
export default enhancer(CheckDetailsDeleteAction);
