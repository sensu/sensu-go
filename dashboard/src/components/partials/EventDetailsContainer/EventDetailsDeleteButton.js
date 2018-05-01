import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { withRouter } from "react-router-dom";
import Button from "material-ui/Button";
import Typography from "material-ui/Typography";
import Dialog, {
  DialogActions,
  DialogContent,
  DialogTitle,
} from "material-ui/Dialog";
import deleteEvent from "/mutations/deleteEvent";

class EventDetailsDeleteButton extends React.PureComponent {
  static propTypes = {
    client: PropTypes.object.isRequired,
    disabled: PropTypes.bool,
    event: PropTypes.object,
    history: PropTypes.object.isRequired,
    onProcessing: PropTypes.func.isRequired,
  };

  static defaultProps = {
    event: null,
    disabled: false,
  };

  static fragments = {
    event: gql`
      fragment EventDetailsDeleteButton_event on Event {
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
  };

  setProcessingTo(newVal) {
    this.props.onProcessing(newVal);
  }

  openDialog = () => {
    this.setState({ dialogOpen: true });
  };

  closeDialog = () => {
    this.setState({ dialogOpen: false });
  };

  deleteEvent = () => {
    const { event, history } = this.props;

    // Cleanup
    this.closeDialog();
    this.setProcessingTo(true);

    // Send request
    const result = deleteEvent(this.props.client, { id: event.id });
    result.then(
      () => {
        this.setProcessingTo(false);
        history.replace(`/${event.ns.org}/${event.ns.env}/events`);
      },
      // eslint-disable-next-line no-console
      err => console.error("error occurred while deleting event", err),
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
          <Button onClick={this.closeDialog} color="primary">
            Cancel
          </Button>
          <Button variant="raised" onClick={this.deleteEvent} color="primary">
            Delete
          </Button>
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

export default withRouter(EventDetailsDeleteButton);
