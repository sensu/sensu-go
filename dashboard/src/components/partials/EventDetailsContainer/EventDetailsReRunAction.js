import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { withRouter } from "react-router-dom";
import Button from "@material-ui/core/Button";
import executeCheck from "/mutations/executeCheck";

class EventDetailsReRunAction extends React.PureComponent {
  static propTypes = {
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
    return <Button onClick={this.handleClick}>Re-Run</Button>;
  }
}

export default withRouter(EventDetailsReRunAction);
