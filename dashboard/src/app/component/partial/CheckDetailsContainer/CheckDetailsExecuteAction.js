import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { compose } from "recompose";
import { withRouter } from "react-router-dom";
import { withApollo } from "react-apollo";

import executeCheck from "/lib/mutation/executeCheck";

import ToastConnector from "/lib/component/relocation/ToastConnector";
import ExecuteCheckStatusToast from "/lib/component/relocation/ExecuteCheckStatusToast";

class CheckDetailsExecuteAction extends React.PureComponent {
  static propTypes = {
    children: PropTypes.func.isRequired,
    client: PropTypes.object.isRequired,
    check: PropTypes.object.isRequired,
  };

  static fragments = {
    check: gql`
      fragment CheckDetailsExecuteAction_check on CheckConfig {
        id
        name
        namespace
      }
    `,
  };

  render() {
    const { children, client, check } = this.props;

    return (
      <ToastConnector>
        {({ addToast }) =>
          children(() => {
            const promise = executeCheck(client, {
              id: check.id,
            });

            addToast(({ remove }) => (
              <ExecuteCheckStatusToast
                onClose={remove}
                mutation={promise}
                checkName={check.name}
                namespace={check.namespace}
              />
            ));
          })
        }
      </ToastConnector>
    );
  }
}

const enhancer = compose(
  withApollo,
  withRouter,
);
export default enhancer(CheckDetailsExecuteAction);
