import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import ClearSilencedEntriesDialog from "/components/partials/ClearSilencedEntriesDialog";

class CheckDetailsClearSilenceAction extends React.PureComponent {
  static propTypes = {
    check: PropTypes.object,
    children: PropTypes.func.isRequired,
    onDone: PropTypes.func.isRequired,
  };

  static defaultProps = {
    check: null,
  };

  static fragments = {
    check: gql`
      fragment CheckDetailsClearSilenceAction_check on CheckConfig {
        id
        isSilenced
        silences {
          ...ClearSilencedEntriesDialog_silence
        }
      }

      ${ClearSilencedEntriesDialog.fragments.silence}
    `,
  };

  state = { isOpen: false };

  render() {
    const { check, children } = this.props;
    const { isOpen } = this.state;

    const open = () => this.setState({ isOpen: true });
    const canOpen = !check.isSilenced;

    return (
      <React.Fragment>
        {children({ canOpen, open })}
        <ClearSilencedEntriesDialog
          silences={check.silences}
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

export default CheckDetailsClearSilenceAction;
