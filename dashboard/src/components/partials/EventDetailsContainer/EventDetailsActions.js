import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import Content from "/components/Content";
import DeleteButton from "./EventDetailsDeleteButton";
import ResolveButton from "./EventDetailsResolveButton";

class EventDetailsActions extends React.PureComponent {
  static propTypes = {
    client: PropTypes.object.isRequired,
    event: PropTypes.object,
    onProcessing: PropTypes.func.isRequired,
  };

  static defaultProps = {
    event: null,
  };

  static fragments = {
    event: gql`
      fragment EventDetailsActions_event on Event {
        ...EventDetailsDeleteButton_event
        ...EventDetailsResolveButton_event
      }

      ${DeleteButton.fragments.event}
      ${ResolveButton.fragments.event}
    `,
  };

  state = {
    locked: false,
  };

  handleProcessing = newVal => {
    this.setState({ locked: newVal });
    this.props.onProcessing(newVal);
  };

  render() {
    const { event, client } = this.props;
    const { locked } = this.state;
    const buttonProps = {
      event,
      client,
      disabled: !event || locked,
      onProcessing: this.handleProcessing,
    };

    return (
      <Content marginBottom>
        <div style={{ flexGrow: 1 }} />
        <ResolveButton {...buttonProps} />
        <DeleteButton {...buttonProps} />
      </Content>
    );
  }
}

export default EventDetailsActions;
