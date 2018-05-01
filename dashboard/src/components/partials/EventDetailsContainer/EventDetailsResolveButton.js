import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import Button from "material-ui/Button";
import resolveEvent from "/mutations/resolveEvent";

class EventDetailsResolveButton extends React.PureComponent {
  static propTypes = {
    client: PropTypes.object.isRequired,
    disabled: PropTypes.bool,
    event: PropTypes.object,
    onProcessing: PropTypes.func.isRequired,
  };

  static defaultProps = {
    disabled: false,
    event: null,
  };

  static fragments = {
    event: gql`
      fragment EventDetailsResolveButton_event on Event {
        id
        check {
          status
        }
      }
    `,
  };

  setProcessingTo(newVal) {
    this.props.onProcessing(newVal);
  }

  resolveEvent = () => {
    this.setProcessingTo(true);

    const result = resolveEvent(this.props.client, { id: this.props.event.id });
    result.then(
      () => this.setProcessingTo(false),
      // eslint-disable-next-line no-console
      err => console.error("error occurred while resolving event", err),
    );
  };

  render() {
    const { disabled, event } = this.props;
    const isDisabled = disabled || event.check.status === 0;

    return (
      <Button onClick={this.resolveEvent} disabled={isDisabled}>
        Resolve
      </Button>
    );
  }
}

export default EventDetailsResolveButton;
