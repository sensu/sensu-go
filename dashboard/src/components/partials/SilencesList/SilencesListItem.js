import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { withProps } from "recompose";

import Avatar from "@material-ui/core/Avatar";
import Button from "@material-ui/core/Button";
import Checkbox from "@material-ui/core/Checkbox";
import Chip from "@material-ui/core/Chip";
import UnsilenceMenuItem from "/components/partials/ToolbarMenuItems/Unsilence";
import Dialog from "@material-ui/core/Dialog";
import DialogActions from "@material-ui/core/DialogActions";
import DialogContent from "@material-ui/core/DialogContent";
import DialogContentText from "@material-ui/core/DialogContentText";
import DialogTitle from "@material-ui/core/DialogTitle";
import FaceIcon from "@material-ui/icons/Face";
import Hidden from "@material-ui/core/Hidden";
import HoverController from "/components/controller/HoverController";
import IconButton from "@material-ui/core/IconButton";
import ModalController from "/components/controller/ModalController";
import NotesIcon from "@material-ui/icons/Notes";
import { RelativeToCurrentDate } from "/components/RelativeDate";
import ResourceDetails from "/components/partials/ResourceDetails";
import SilenceExpiration from "/components/partials/SilenceExpiration";
import Slide from "@material-ui/core/Slide";
import TableCell from "@material-ui/core/TableCell";
import TableOverflowCell from "/components/partials/TableOverflowCell";
import TableSelectableRow from "/components/partials/TableSelectableRow";
import { FloatingTableToolbarCell } from "/components/partials/TableToolbarCell";
import ToolbarMenu from "/components/partials/ToolbarMenu";
import Tooltip from "@material-ui/core/Tooltip";

const SlideUp = withProps({ direction: "up" })(Slide);
const RightAlign = withProps({
  style: {
    display: "flex",
    justifyContent: "flex-end",
  },
})("div");

class SilencesListItem extends React.Component {
  static propTypes = {
    editable: PropTypes.bool.isRequired,
    editing: PropTypes.bool.isRequired,
    silence: PropTypes.object.isRequired,
    hovered: PropTypes.bool.isRequired,
    onHover: PropTypes.func.isRequired,
    selected: PropTypes.bool.isRequired,
    onClickClearSilences: PropTypes.func.isRequired,
    onClickSelect: PropTypes.func.isRequired,
  };

  static fragments = {
    silence: gql`
      fragment SilencesListItem_silence on Silenced {
        ...SilenceExpiration_silence
        name
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
            <RelativeToCurrentDate dateTime={silence.begin} />
          </strong>
          .
        </React.Fragment>
      );
    }

    return <SilenceExpiration silence={silence} />;
  };

  render() {
    const { editable, editing, silence, selected, onClickSelect } = this.props;

    return (
      <HoverController onHover={this.props.onHover}>
        <TableSelectableRow selected={selected}>
          {editable && (
            <TableCell padding="checkbox">
              <Checkbox
                checked={selected}
                onChange={() => onClickSelect(!selected)}
              />
            </TableCell>
          )}

          <TableOverflowCell>
            <ResourceDetails
              title={<strong>{silence.name}</strong>}
              details={this.renderDetails()}
            />
          </TableOverflowCell>

          <Hidden only="xs">
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
          </Hidden>

          <FloatingTableToolbarCell
            hovered={this.props.hovered}
            disabled={!editable || editing}
          >
            {() => (
              <RightAlign>
                {silence.reason && (
                  <ModalController
                    renderModal={({ close }) => (
                      <Dialog
                        open
                        fullWidth
                        TransitionComponent={SlideUp}
                        onClose={close}
                      >
                        <DialogTitle>Reason Given</DialogTitle>
                        <DialogContent>
                          <DialogContentText>
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

                <ToolbarMenu>
                  <ToolbarMenu.Item id="delete" visible="never">
                    <UnsilenceMenuItem
                      onClick={this.props.onClickClearSilences}
                    />
                  </ToolbarMenu.Item>
                </ToolbarMenu>
              </RightAlign>
            )}
          </FloatingTableToolbarCell>
        </TableSelectableRow>
      </HoverController>
    );
  }
}

export default SilencesListItem;
