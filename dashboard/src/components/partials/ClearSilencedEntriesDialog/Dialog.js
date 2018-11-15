import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { withApollo } from "react-apollo";
import { compose, hoistStatics } from "recompose";
import { withStyles } from "@material-ui/core/styles";
import {
  darken,
  fade,
  lighten,
} from "@material-ui/core/styles/colorManipulator";

import Button from "@material-ui/core/Button";
import Checkbox from "@material-ui/core/Checkbox";
import Dialog from "@material-ui/core/Dialog";
import DialogActions from "@material-ui/core/DialogActions";
import DialogContent from "@material-ui/core/DialogContent";
import DialogContentText from "@material-ui/core/DialogContentText";
import DialogContentParagraph from "/components/DialogContentParagraph";
import DialogTitle from "@material-ui/core/DialogTitle";
import ListController from "/components/controller/ListController";
import Loader from "/components/util/Loader";
import ResourceDetails from "/components/partials/ResourceDetails";
import SilenceExpiration from "/components/partials/SilenceExpiration";
import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableCell from "@material-ui/core/TableCell";
import TableOverflowCell from "/components/partials/TableOverflowCell";
import TableSelectableRow from "/components/partials/TableSelectableRow";
import withMobileDialog from "@material-ui/core/withMobileDialog";
import deleteSilence from "/mutations/deleteSilence";

const StyledTable = withStyles(theme => ({
  root: {
    // https://github.com/mui-org/material-ui/blob/a207808/packages/material-ui/src/TableCell/TableCell.js#L13-L14
    borderTop: `1px solid ${
      theme.palette.type === "light"
        ? lighten(fade(theme.palette.divider, 1), 0.88)
        : darken(fade(theme.palette.divider, 1), 0.8)
    }`,
  },
}))(Table);

class ClearSilencedEntriesDialog extends React.PureComponent {
  static propTypes = {
    client: PropTypes.object.isRequired,
    close: PropTypes.func.isRequired,
    fullScreen: PropTypes.bool.isRequired,
    open: PropTypes.bool,
    silences: PropTypes.array,
  };

  static defaultProps = {
    open: false,
    silences: null,
  };

  static fragments = {
    silence: gql`
      fragment ClearSilencedEntriesDialog_silence on Silenced {
        ...SilenceExpiration_silence

        id
        deleted @client
        name
        creator {
          username
        }
      }

      ${SilenceExpiration.fragments.silence}
    `,
  };

  state = {
    submitting: false,
  };

  clearItems = items => {
    const { client, close } = this.props;
    const done = () => this.setState({ submitting: false });
    const clear = ({ id }) => deleteSilence(client, { id });

    this.setState({ submitting: true });
    Promise.all(items.map(clear))
      .then(done)
      .then(close)
      .catch(done);
  };

  renderListItem = ({ key, item: silence, selected, setSelected }) => (
    <TableSelectableRow
      selected={selected}
      key={key}
      style={{ verticalAlign: "middle", cursor: "pointer" }}
      onClick={() => setSelected(!selected)}
    >
      <TableCell padding="checkbox">
        <Checkbox
          color="primary"
          checked={selected}
          onChange={e => setSelected(e.target.checked)}
        />
      </TableCell>
      <TableOverflowCell>
        <ResourceDetails
          title={silence.name}
          details={<SilenceExpiration silence={silence} />}
        />
      </TableOverflowCell>
      <TableCell>
        <ResourceDetails title={silence.creator.username} />
      </TableCell>
    </TableSelectableRow>
  );

  renderEmpty = () => (
    <DialogContent>
      <DialogContentParagraph>
        {`There doesn't seem to be anything here. This may can occur when
        the silence(s) have already been cleared or have expired.`}
      </DialogContentParagraph>
    </DialogContent>
  );

  render() {
    const { open, close, fullScreen, silences: silencesProp } = this.props;
    const { submitting } = this.state;

    // omit duplicates
    const silences = Object.values(
      (silencesProp || [])
        .filter(sl => !sl.deleted)
        .reduce((memo, sl) => Object.assign({ [sl.name]: sl }, memo), {}),
    );

    return (
      <Dialog fullWidth fullScreen={fullScreen} open={open} onClose={close}>
        <ListController
          items={silences}
          getItemKey={node => node.name}
          renderEmptyState={this.renderEmpty}
          renderItem={this.renderListItem}
        >
          {({ children, selectedItems }) => (
            <Loader loading={submitting} passthrough>
              <DialogTitle>Clear Silencing Entries</DialogTitle>
              <DialogContent style={{ paddingBottom: 8 }}>
                <DialogContentText>
                  Select all entries you would like to clear.
                </DialogContentText>
              </DialogContent>
              <DialogContent style={{ paddingLeft: 0, paddingRight: 0 }}>
                <StyledTable>
                  <TableBody>{children}</TableBody>
                </StyledTable>
              </DialogContent>
              <DialogActions>
                <Button onClick={close} color="default">
                  Cancel
                </Button>
                <Button
                  onClick={() => this.clearItems(selectedItems)}
                  color="primary"
                  variant="raised"
                  autoFocus
                  disabled={selectedItems.length === 0 || submitting}
                >
                  Clear
                </Button>
              </DialogActions>
            </Loader>
          )}
        </ListController>
      </Dialog>
    );
  }
}

const enhancer = compose(withApollo, withMobileDialog({ breakpoint: "xs" }));
export default hoistStatics(enhancer)(ClearSilencedEntriesDialog);
