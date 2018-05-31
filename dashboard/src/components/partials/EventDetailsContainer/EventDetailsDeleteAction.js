import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { withRouter } from "react-router-dom";
import Button from "@material-ui/core/Button";
import Typography from "@material-ui/core/Typography";
import Dialog from "@material-ui/core/Dialog";
import DialogActions from "@material-ui/core/DialogActions";
import DialogContent from "@material-ui/core/DialogContent";
import DialogTitle from "@material-ui/core/DialogTitle";
import ButtonSet from "/components/ButtonSet";
import deleteEvent from "/mutations/deleteEvent";

class EventDetailsDeleteAction extends React.PureComponent {
  static propTypes = {
    client: PropTypes.object.isRequired,
    disabled: PropTypes.bool,
    event: PropTypes.object,
    history: PropTypes.object.isRequired,
    onRequestStart: PropTypes.func.isRequired,
    onRequestEnd: PropTypes.func.isRequired,
  };

  static defaultProps = {
    event: null,
    disabled: false,
  };

  static fragments = {
    event: gql`
      fragment EventDetailsDeleteAction_event on Event {
        id
        ns: namespace {
          org: organization
          env: environment
        }
      }
    `,
  };

  state = {
    dialogOpen: false,
    locked: false,
  };

  requestStart() {
    this.props.onRequestStart();
    this.setState({ locked: true });
  }

  requestEnd() {
    this.props.onRequestEnd();
    this.setState({ locked: false });
  }

  openDialog = () => {
    this.setState({ dialogOpen: true });
  };

  closeDialog = () => {
    this.setState({ dialogOpen: false });
  };

  deleteEvent = () => {
    const {
      client,
      event: { id, ns },
      history,
    } = this.props;
    if (this.state.locked) {
      return;
    }

    // Cleanup
    this.closeDialog();
    this.requestStart();

    // Send request
    deleteEvent(client, { id }).then(
      () => {
        this.requestEnd();
        history.replace(`/${ns.org}/${ns.env}/events`);
      },
      error => {
        this.requestEnd();
        throw error;
      },
    );
  };

  _renderDialog = () => {
    const { dialogOpen } = this.state;
    return (
      <Dialog
        disableBackdropClick
        disableEscapeKeyDown
        maxWidth="xs"
        open={dialogOpen}
        aria-labelledby="confirmation-dialog-title"
      >
        <DialogTitle id="confirmation-dialog-title">Confirmation</DialogTitle>
        <DialogContent>
          <Typography>
            Are you sure you would like to delete this event?
          </Typography>
        </DialogContent>
        <DialogActions>
          <ButtonSet>
            <Button onClick={this.closeDialog} color="primary">
              Cancel
            </Button>
            <Button variant="raised" onClick={this.deleteEvent} color="primary">
              Delete
            </Button>
          </ButtonSet>
        </DialogActions>
      </Dialog>
    );
  };

  _renderButton = () => {
    const { disabled } = this.props;
    return (
      <Button variant="raised" onClick={this.openDialog} disabled={disabled}>
        Delete
      </Button>
    );
  };

  render() {
    return (
      <React.Fragment>
        {this._renderDialog()}
        {this._renderButton()}
      </React.Fragment>
    );
  }
}

export default withRouter(EventDetailsDeleteAction);
