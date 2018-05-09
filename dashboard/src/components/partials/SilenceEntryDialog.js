import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "material-ui/styles";

import Button from "material-ui/Button";
import Dialog, {
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  withMobileDialog,
} from "material-ui/Dialog";

import Loader from "/components/util/Loader";

import SilenceEntryForm from "/components/partials/SilenceEntryForm";

const StyledDialogContentText = withStyles(() => ({
  root: { marginBottom: "2rem" },
}))(DialogContentText);

class SilenceEntryDialog extends React.PureComponent {
  static propTypes = {
    // fullScreen prop is controlled by the `withMobileDialog` enhancer.
    fullScreen: PropTypes.bool.isRequired,
    onClose: PropTypes.func,
    onSave: PropTypes.func,
    loading: PropTypes.bool,
    values: PropTypes.object,
  };

  static defaultProps = {
    onClose: undefined,
    onSave: undefined,
    loading: undefined,
    values: {},
  };

  formRef = React.createRef();

  render() {
    const { fullScreen, loading, onClose, onSave, values } = this.props;

    return (
      <Dialog open fullScreen={fullScreen} onClose={onClose}>
        <Loader loading={loading} passthrough>
          <DialogTitle>
            {values.id ? "Edit Silencing Entry" : "New Silencing Entry"}
          </DialogTitle>
          <DialogContent>
            <StyledDialogContentText>
              Create a silencing entry to temporarily prevent check result
              handlers from being triggered. A full reference to check silencing
              is available on the Sensu docs site.<br />
              <a
                href="https://docs.sensu.io/sensu-core/2.0/reference/silencing/"
                target="_docs"
              >
                Learn more
              </a>
            </StyledDialogContentText>
            <SilenceEntryForm
              ref={this.formRef}
              values={values}
              onSubmit={onSave}
            />
          </DialogContent>
          <DialogActions>
            <Button onClick={onClose} color="primary">
              Cancel
            </Button>
            <Button
              onClick={() => {
                this.formRef.current.submit();
              }}
              color="primary"
              variant="raised"
              autoFocus
            >
              {values.id ? "Save" : "Create"}
            </Button>
          </DialogActions>
        </Loader>
      </Dialog>
    );
  }
}
export default withMobileDialog({ breakpoint: "xs" })(SilenceEntryDialog);
