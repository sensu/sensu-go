import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "@material-ui/core/styles";
import Button from "@material-ui/core/Button";
import ButtonSet from "/components/ButtonSet";
import withMobileDialog from "@material-ui/core/withMobileDialog";
import Dialog from "@material-ui/core/Dialog";
import DialogActions from "@material-ui/core/DialogActions";
import DialogContent from "@material-ui/core/DialogContent";
import DialogTitle from "@material-ui/core/DialogTitle";
import Typography from "@material-ui/core/Typography";
import List from "@material-ui/core/List";
import ListItem from "@material-ui/core/ListItem";
import ListItemText from "@material-ui/core/ListItemText";
import ListItemSecondaryAction from "@material-ui/core/ListItemSecondaryAction";
import Checkbox from "@material-ui/core/Checkbox";

const Highlight = withStyles({
  root: {
    display: "inline",
    fontWeight: 600,
  },
})(props => <Typography component="strong" {...props} />);

const styles = {
  resourceList: {
    margin: "20px 0",
    border: "1px solid grey",
    maxHeight: "150px",
    overflowY: "scroll",
    whiteSpace: "nowrap",
    overflowX: "hidden",
    textOverflow: "ellipsis",
  },
};

class ConfirmUnsilenceDialog extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    fullScreen: PropTypes.bool.isRequired,
    identifier: PropTypes.node,
    open: PropTypes.bool.isRequired,
    onConfirm: PropTypes.func.isRequired,
    onClose: PropTypes.func.isRequired,
    resources: PropTypes.array,
  };

  static defaultProps = {
    identifier: "this resource",
    resources: [],
  };

  state = {
    selectedItems: this.props.resources,
  };

  handleToggle = resource => () => {
    const newState = this.state.selectedItems;
    if (newState.some(r => r.id === resource.id)) {
      const index = newState.indexOf(resource);
      newState.splice(index, 1);
    } else {
      newState.push(resource);
    }
    this.setState({ selectedItems: newState });
  };

  render() {
    const { classes, resources } = this.props;
    const titleId = "confirm-unsilence-dialog-title";
    const these =
      resources.length <= 1 ? this.props.identifier : "these resources";
    const resourceList = resources.map(resource => (
      <ListItem key={resource.id}>
        <ListItemText primary={resource.id} />
        <ListItemSecondaryAction>
          <Checkbox
            onChange={this.handleToggle(resource)}
            checked={this.state.selectedItems.some(r => r.id === resource.id)}
          />
        </ListItemSecondaryAction>
      </ListItem>
    ));

    return (
      <Dialog
        aria-labelledby={titleId}
        disableBackdropClick
        disableEscapeKeyDown
        fullScreen={this.props.fullScreen}
        maxWidth="sm"
        open={this.props.open}
      >
        <DialogTitle id={titleId}>Confirm</DialogTitle>
        <DialogContent>
          <Typography>
            Are you sure you would like to clear silences for{" "}
            <Highlight>{these}</Highlight>?
          </Typography>
          {!resourceList.empty && (
            <div className={classes.resourceList}>
              <List dense>{resourceList}</List>
            </div>
          )}
          <Typography>This will enable notifications again.</Typography>
        </DialogContent>
        <DialogActions>
          <ButtonSet>
            <Button onClick={this.props.onClose} color="default">
              Cancel
            </Button>
            <Button
              variant="raised"
              onClick={() => this.props.onConfirm(this.state.selectedItems)}
              color="primary"
            >
              Clear Silence
            </Button>
          </ButtonSet>
        </DialogActions>
      </Dialog>
    );
  }
}

const enhancer = withMobileDialog({ breakpoint: "xs" });
export default withStyles(styles)(enhancer(ConfirmUnsilenceDialog));
