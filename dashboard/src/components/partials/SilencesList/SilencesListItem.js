import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Avatar from "@material-ui/core/Avatar";
import Button from "@material-ui/core/Button";
import Checkbox from "@material-ui/core/Checkbox";
import Chip from "@material-ui/core/Chip";
import DisclosureIcon from "@material-ui/icons/MoreVert";
import Dialog from "@material-ui/core/Dialog";
import DialogActions from "@material-ui/core/DialogActions";
import DialogContent from "@material-ui/core/DialogContent";
import DialogContentText from "@material-ui/core/DialogContentText";
import DialogTitle from "@material-ui/core/DialogTitle";
import FaceIcon from "@material-ui/icons/Face";
import IconButton from "@material-ui/core/IconButton";
import Menu from "@material-ui/core/Menu";
import MenuController from "/components/controller/MenuController";
import MenuItem from "@material-ui/core/MenuItem";
import ModalController from "/components/controller/ModalController";
import NotesIcon from "@material-ui/icons/Notes";
import RelativeDate from "/components/RelativeDate";
import ResourceDetails from "/components/partials/ResourceDetails";
import RootRef from "@material-ui/core/RootRef";
import SilenceExpiration from "/components/partials/SilenceExpiration";
import Slide from "@material-ui/core/Slide";
import TableCell from "@material-ui/core/TableCell";
import TableOverflowCell from "/components/partials/TableOverflowCell";
import TableSelectableRow from "/components/partials/TableSelectableRow";
import Tooltip from "@material-ui/core/Tooltip";

class SilencesListItem extends React.Component {
  static propTypes = {
    silence: PropTypes.object.isRequired,
    selected: PropTypes.bool.isRequired,
    onClickSelect: PropTypes.func.isRequired,
    onClickDelete: PropTypes.func.isRequired,
  };

  static fragments = {
    silence: gql`
      fragment SilencesListItem_silence on Silenced {
        ...SilenceExpiration_silence
        storeId
        begin
        reason
        creator {
          username
        }
      }

      ${SilenceExpiration.fragments.silence}
    `,
  };

  renderDetails = () => {
    const { silence } = this.props;

    if (new Date(silence.begin) > new Date()) {
      return (
        <React.Fragment>
          Takes effect{" "}
          <strong>
            <RelativeDate dateTime={silence.begin} />
          </strong>.
        </React.Fragment>
      );
    }

    return <SilenceExpiration silence={silence} />;
  };

  render() {
    const { silence, selected, onClickSelect, onClickDelete } = this.props;

    return (
      <TableSelectableRow selected={selected}>
        <TableCell padding="checkbox">
          <Checkbox
            checked={selected}
            onChange={() => onClickSelect(!selected)}
          />
        </TableCell>
        <TableOverflowCell>
          <ResourceDetails
            title={<strong>{silence.storeId}</strong>}
            details={this.renderDetails()}
          />
        </TableOverflowCell>
        <TableCell
          padding="none"
          style={{
            // TODO: magic number
            paddingTop: 8, // one spacing unit
          }}
        >
          <Chip
            avatar={
              <Avatar>
                <FaceIcon />
              </Avatar>
            }
            label={silence.creator.username}
            style={{
              // TODO: ideally have Chip scale to current fontSize(?)
              transform: "scale(0.87)",
            }}
          />
        </TableCell>
        <TableCell padding="checkbox">
          {silence.reason && (
            <ModalController
              renderModal={({ close }) => (
                <Dialog
                  open
                  fullWidth
                  TransitionComponent={props => (
                    <Slide direction="up" {...props} />
                  )}
                  onClose={close}
                >
                  <DialogTitle>Silenced For...</DialogTitle>
                  <DialogContent>
                    <DialogContentText id="alert-dialog-slide-description">
                      {silence.reason}
                    </DialogContentText>
                  </DialogContent>
                  <DialogActions>
                    <Button onClick={close} color="contrast">
                      Close
                    </Button>
                  </DialogActions>
                </Dialog>
              )}
            >
              {({ open }) => (
                <Tooltip title={"Reason"}>
                  <IconButton onClick={open}>
                    <NotesIcon />
                  </IconButton>
                </Tooltip>
              )}
            </ModalController>
          )}
        </TableCell>
        <TableCell padding="checkbox">
          <MenuController
            renderMenu={({ anchorEl, close }) => (
              <Menu open onClose={close} anchorEl={anchorEl}>
                <MenuItem
                  onClick={() => {
                    onClickDelete();
                    this.closeMenu();
                  }}
                >
                  Delete
                </MenuItem>
              </Menu>
            )}
          >
            {({ open, ref }) => (
              <RootRef rootRef={ref}>
                <IconButton onClick={open}>
                  <DisclosureIcon />
                </IconButton>
              </RootRef>
            )}
          </MenuController>
        </TableCell>
      </TableSelectableRow>
    );
  }
}

export default SilencesListItem;
