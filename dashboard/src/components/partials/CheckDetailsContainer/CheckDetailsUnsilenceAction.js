import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import ClearSilencedEntriesDialog from "/components/partials/ClearSilencedEntriesDialog";

class CheckDetailsUnsilenceAction extends React.PureComponent {
  static propTypes = {
    check: PropTypes.object,
    children: PropTypes.func.isRequired,
    refetch: PropTypes.func.isRequired,
  };

  static defaultProps = {
    check: null,
  };

  static fragments = {
    check: gql`
      fragment CheckDetailsUnsilenceAction_check on CheckConfig {
        id
        silences {
          ...ClearSilencedEntriesDialog_silence
        }
      }

      ${ClearSilencedEntriesDialog.fragments.silence}
    `,
  };

  state = {
    unsilence: null,
  };

  openDialog = () => {
    this.setState({
      unsilence: this.props.check.silences,
    });
  };

  closeDialog = () => {
    this.setState({ unsilence: null });
  };

  render() {
    const { children, refetch } = this.props;
    const { unsilence } = this.state;

    return (
      <React.Fragment>
        <ClearSilencedEntriesDialog
          silences={unsilence}
          open={!!unsilence}
          close={() => {
            this.closeDialog();
            refetch();
          }}
        />
        {children(this.openDialog)}
      </React.Fragment>
    );
  }
}

export default CheckDetailsUnsilenceAction;
