import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import Button from "material-ui/Button";
import resolveEvent from "/mutations/resolveEvent";

class EventDetailsResolveAction extends React.PureComponent {
  static propTypes = {
    client: PropTypes.object.isRequired,
    disabled: PropTypes.bool,
    event: PropTypes.object,
  };

  static defaultProps = {
    disabled: false,
    event: null,
  };

  static fragments = {
    event: gql`
      fragment EventDetailsResolveAction_event on Event {
        id
        check {
          status
        }
      }
    `,
  };

  state = {
    locked: false,
  };

  resolveStart() {
    this.setState({ locked: true });
  }

  resolveEnd() {
    this.setState({ locked: false });
  }

  resolveEvent = () => {
    const {
      client,
      event: { id },
    } = this.props;
    if (this.state.locked) return;

    this.resolveStart();
    resolveEvent(client, { id }).then(
      () => this.resolveEnd(),
      error => {
        this.resolveEnd();
        throw error;
      },
    );
  };

  render() {
    const { disabled: disabledProp, event, ...props } = this.props;
    const disabled = disabledProp || event.check.status === 0;

    return (
      <Button onClick={this.resolveEvent} disabled={disabled} {...props}>
        Resolve
      </Button>
    );
  }
}

export default EventDetailsResolveAction;
