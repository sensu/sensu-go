import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { compose } from "recompose";
import { withRouter } from "react-router-dom";
import { withApollo } from "react-apollo";

import executeCheck from "/mutations/executeCheck";

import ToastConnector from "/components/relocation/ToastConnector";
import ExecuteCheckStatusToast from "/components/relocation/ExecuteCheckStatusToast";

class EventDetailsReRunAction extends React.PureComponent {
  static propTypes = {
    children: PropTypes.func.isRequired,
    client: PropTypes.object.isRequired,
    event: PropTypes.object.isRequired,
  };

  static fragments = {
    event: gql`
      fragment EventDetailsReRunAction_event on Event {
        check {
          nodeId
        }
        entity {
          name
        }
        namespace
      }
    `,
  };

  render() {
    const { children, event, client } = this.props;

    return (
      <ToastConnector>
        {({ addToast }) =>
          children(() => {
            const promise = executeCheck(client, {
              id: event.check.nodeId,
              subscriptions: [`entity:${event.entity.name}`],
            });

            addToast(({ remove }) => (
              <ExecuteCheckStatusToast
                onClose={remove}
                mutation={promise}
                checkName={event.check.name}
                entityName={event.entity.name}
                namespace={event.namespace}
              />
            ));
          })
        }
      </ToastConnector>
    );
  }
}

const enhancer = compose(withApollo, withRouter);
export default enhancer(EventDetailsReRunAction);
