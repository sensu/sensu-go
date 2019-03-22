import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { withApollo } from "react-apollo";

import setCheckPublish from "/lib/mutation/setCheckPublish";

import ToastConnector from "/lib/component/relocation/ToastConnector";
import PublishCheckStatusToast from "/lib/component/relocation/PublishCheckStatusToast";

class CheckDetailsUnpublishAction extends React.PureComponent {
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
              publish: false,
            });
            addToast(({ remove }) => (
              <PublishCheckStatusToast
                onClose={remove}
                mutation={promise}
                checkName={check.name}
                publish={false}
              />
            ));
          })
        }
      </ToastConnector>
    );
  }
}

export default withApollo(CheckDetailsUnpublishAction);
