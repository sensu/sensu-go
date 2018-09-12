import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import ClearSilencedEntriesDialog from "/components/partials/ClearSilencedEntriesDialog";

class EntityDetailsClearSilenceAction extends React.Component {
  static propTypes = {
    children: PropTypes.func.isRequired,
    entity: PropTypes.object.isRequired,
    onDone: PropTypes.func,
  };

  static defaultProps = {
    onDone: () => false,
  };

  static fragments = {
    entity: gql`
      fragment EntityDetailsClearSilenceAction_entity on Entity {
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
    const { entity } = this.props;
    const { isOpen } = this.state;

    const canOpen = !entity.isSilenced;
    const open = () => this.setState({ isOpen: true });

    return (
      <React.Fragment>
        {this.props.children({ canOpen, open })}
        <ClearSilencedEntriesDialog
          silences={entity.silences}
          open={isOpen}
          close={() => {
            this.setState({ isOpen: false });
            this.props.onDone();
          }}
        />
      </React.Fragment>
    );
  }
}

export default EntityDetailsClearSilenceAction;
