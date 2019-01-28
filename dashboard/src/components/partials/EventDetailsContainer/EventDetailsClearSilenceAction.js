import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import ClearSilencedEntriesDialog from "/components/partials/ClearSilencedEntriesDialog";

class EventDetailsClearSilenceAction extends React.PureComponent {
  static propTypes = {
    event: PropTypes.object,
    children: PropTypes.func.isRequired,
    onDone: PropTypes.func.isRequired,
  };

  static defaultProps = {
    event: null,
  };

  static fragments = {
    event: gql`
      fragment EventDetailsClearSilenceAction_event on Event {
        check {
          silenced
          silences {
            ...ClearSilencedEntriesDialog_silence
          }
        }
      }

      ${ClearSilencedEntriesDialog.fragments.silence}
    `,
  };

  state = { isOpen: false };

  render() {
    const { event, children } = this.props;
    const { isOpen } = this.state;

    const open = () => this.setState({ isOpen: true });
    const canOpen = event.check.silenced.length > 0;

    return (
      <React.Fragment>
        {children({ canOpen, open })}
        <ClearSilencedEntriesDialog
          silences={event.check.silences}
          open={isOpen}
          close={() => {
            this.props.onDone();
            this.setState({ isOpen: false });
          }}
        />
      </React.Fragment>
    );
  }
}

export default EventDetailsClearSilenceAction;
