import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { compose } from "recompose";
import { withRouter } from "react-router-dom";
import { withApollo } from "react-apollo";

import executeCheck from "/mutations/executeCheck";

class EventDetailsReRunAction extends React.PureComponent {
  static propTypes = {
    children: PropTypes.func.isRequired,
    client: PropTypes.object.isRequired,
    event: PropTypes.object,
  };

  static defaultProps = {
    event: null,
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
      }
    `,
  };

  handleClick = () => {
    const {
      client,
      event: { check, entity },
    } = this.props;

    executeCheck(client, {
      id: check.nodeId,
      subscriptions: [`entity:${entity.name}`],
    });
  };

  render() {
    return this.props.children(this.handleClick);
  }
}

const enhancer = compose(withApollo, withRouter);
export default enhancer(EventDetailsReRunAction);
