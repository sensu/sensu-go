import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { withApollo } from "react-apollo";

import setCheckPublish from "/mutations/setCheckPublish";

import ToastConnector from "/components/relocation/ToastConnector";
import PublishCheckStatusToast from "/components/relocation/PublishCheckStatusToast";

class CheckDetailsPublishAction extends React.PureComponent {
  static propTypes = {
    children: PropTypes.func.isRequired,
    client: PropTypes.object.isRequired,
    check: PropTypes.object.isRequired,
  };

  static fragments = {
    check: gql`
      fragment CheckDetailsPublishAction_check on CheckConfig {
        id
        name
        publish
      }
    `,
  };

  render() {
    const { children, client, check } = this.props;

    return (
      <ToastConnector>
        {({ addToast }) =>
          children(() => {
            const promise = setCheckPublish(client, {
              id: check.id,
              publish: true,
            });
            addToast(({ remove }) => (
              <PublishCheckStatusToast
                onClose={remove}
                mutation={promise}
                checkName={check.name}
              />
            ));
          })
        }
      </ToastConnector>
    );
  }
}

export default withApollo(CheckDetailsPublishAction);
