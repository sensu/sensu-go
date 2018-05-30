import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "@material-ui/core/styles";
import Dialog from "@material-ui/core/Dialog";
import DialogActions from "@material-ui/core/DialogActions";
import DialogContent from "@material-ui/core/DialogContent";
import DialogTitle from "@material-ui/core/DialogTitle";
import Typography from "@material-ui/core/Typography";
import Button from "@material-ui/core/Button";
import ButtonSet from "/components/ButtonSet";

const Highlight = withStyles({
  root: {
    display: "inline",
    fontWeight: 600,
  },
})(props => <Typography component="strong" {...props} />);

class ConfirmDeleteDialog extends React.Component {
  static propTypes = {
    open: PropTypes.bool.isRequired,
    onConfirm: PropTypes.func.isRequired,
    onClose: PropTypes.func.isRequired,
    identifier: PropTypes.node,
  };

  static defaultProps = {
    identifier: "this resource",
  };

  render() {
    return (
      <Dialog
        disableBackdropClick
        disableEscapeKeyDown
        maxWidth="sm"
        open={this.props.open}
        aria-labelledby="confirmation-dialog-title"
      >
        <DialogTitle id="confirmation-dialog-title">Confirm</DialogTitle>
        <DialogContent>
          <Typography>
            Are you sure you would like to permanantly delete{" "}
            <Highlight>{this.props.identifier}</Highlight>? This operation{" "}
            <Highlight>cannot</Highlight> be undone.
          </Typography>
        </DialogContent>
        <DialogActions>
          <ButtonSet>
            <Button onClick={this.props.onClose} color="default">
              Cancel
            </Button>
            <Button
              variant="raised"
              onClick={this.props.onConfirm}
              color="primary"
            >
              Delete
            </Button>
          </ButtonSet>
        </DialogActions>
      </Dialog>
    );
  }
}

export default ConfirmDeleteDialog;
