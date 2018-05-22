import React from "react";
import PropTypes from "prop-types";
import { connect } from "react-redux";
import { compose } from "lodash/fp";
import { withStyles } from "@material-ui/core/styles";

import Modal from "@material-ui/core/Modal";
import Paper from "@material-ui/core/Paper";
import List from "@material-ui/core/List";
import ListItem from "@material-ui/core/ListItem";
import ListItemIcon from "@material-ui/core/ListItemIcon";
import ListItemSecondaryAction from "@material-ui/core/ListItemSecondaryAction";
import ListItemText from "@material-ui/core/ListItemText";
import ListSubheader from "@material-ui/core/ListSubheader";
import MenuItem from "@material-ui/core/MenuItem";
import Menu from "@material-ui/core/Menu";
import Switch from "@material-ui/core/Switch";
import BulbIcon from "@material-ui/icons/LightbulbOutline";
import EyeIcon from "@material-ui/icons/RemoveRedEye";

const styles = theme => ({
  root: {
    alignItems: "center",
    justifyContent: "center",
    backdropFilter: "blur(3px)",
  },
  paper: {
    minWidth: 360,
    padding: theme.spacing.unit * 2,
  },
});

class Preferences extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    theme: PropTypes.shape({
      theme: PropTypes.string,
      dark: PropTypes.bool,
    }).isRequired,
    open: PropTypes.bool.isRequired,
    onClose: PropTypes.func.isRequired,
    dispatch: PropTypes.func.isRequired,
  };

  state = {
    anchorEl: null,
  };

  handleToggle = () => {
    this.props.dispatch({ type: "theme/TOGGLE_DARK_MODE" });
  };

  handleThemeSelect = theme => () => {
    this.props.dispatch({
      type: "theme/CHANGE",
      payload: { theme },
    });
    this.setState({ anchorEl: null });
  };

  handleThemeClick = event => {
    this.setState({ anchorEl: event.currentTarget });
  };

  handleThemeClose = () => {
    this.setState({ anchorEl: null });
  };

  render() {
    const { classes, open, onClose, theme } = this.props;
    const { anchorEl } = this.state;

    return (
      <Modal className={classes.root} open={open} onClose={onClose}>
        <Paper className={classes.paper}>
          <List subheader={<ListSubheader>Appearance</ListSubheader>}>
            <ListItem>
              <ListItemIcon>
                <BulbIcon />
              </ListItemIcon>
              <ListItemText
                primary="Lights Out"
                secondary="Switch to the dark theme..."
              />
              <ListItemSecondaryAction>
                <Switch onChange={this.handleToggle} checked={theme.dark} />
              </ListItemSecondaryAction>
            </ListItem>
            <ListItem button onClick={this.handleThemeClick}>
              <ListItemIcon>
                <EyeIcon />
              </ListItemIcon>
              <ListItemText primary="Theme" secondary={theme.theme} />
            </ListItem>
          </List>
          <Menu
            id="theme-menu"
            anchorEl={anchorEl}
            open={Boolean(anchorEl)}
            onClose={this.handleThemeClose}
          >
            <MenuItem onClick={this.handleThemeSelect("sensu")}>
              Default
            </MenuItem>
            <MenuItem onClick={this.handleThemeSelect("classic")}>
              Classic
            </MenuItem>
            <MenuItem onClick={this.handleThemeSelect("uchiwa")}>
              Uchiwa
            </MenuItem>
            <MenuItem onClick={this.handleThemeSelect("dva")}>DVA</MenuItem>
          </Menu>
        </Paper>
      </Modal>
    );
  }
}

const enhancer = compose(
  connect(st => ({ theme: st.theme })),
  withStyles(styles),
);
export default enhancer(Preferences);
