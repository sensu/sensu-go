import React from "react";
import PropTypes from "prop-types";
import { compose } from "recompose";
import { withApollo } from "react-apollo";

import Button from "@material-ui/core/Button";
import Dialog from "@material-ui/core/Dialog";
import DialogActions from "@material-ui/core/DialogActions";
import DialogContent from "@material-ui/core/DialogContent";
import DialogContentText from "@material-ui/core/DialogContentText";
import DialogTitle from "@material-ui/core/DialogTitle";
import withMobileDialog from "@material-ui/core/withMobileDialog";

import invalidateTokens from "/lib/mutation/invalidateTokens";

class AuthInvalidDialog extends React.PureComponent {
  static propTypes = {
    // fullScreen prop is controlled by the `withMobileDialog` enhancer.
    fullScreen: PropTypes.bool.isRequired,
    client: PropTypes.object.isRequired,
  };

  render() {
    const { fullScreen, client } = this.props;

    return (
      <Dialog open fullScreen={fullScreen}>
        <DialogTitle>Session Expired</DialogTitle>
        <DialogContent>
          <DialogContentText>
            Your session has expired. Please sign in to continue.
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button
            onClick={() => {
              invalidateTokens(client);
            }}
            color="primary"
          >
            Sign In
          </Button>
        </DialogActions>
      </Dialog>
    );
  }
}

export default compose(
  withMobileDialog({ breakpoint: "xs" }),
  withApollo,
)(AuthInvalidDialog);
