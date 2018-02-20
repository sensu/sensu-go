import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "material-ui/styles";

import Modal from "material-ui/Modal";
import Paper from "material-ui/Paper";
import List, {
  ListItem,
  ListItemIcon,
  ListItemSecondaryAction,
  ListItemText,
  ListSubheader,
} from "material-ui/List";
import Menu, { MenuItem } from "material-ui/Menu";
import Switch from "material-ui/Switch";
import BulbIcon from "material-ui-icons/LightbulbOutline";
import EyeIcon from "material-ui-icons/RemoveRedEye";

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
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    open: PropTypes.bool.isRequired,
    onClose: PropTypes.func.isRequired,
  };

  static contextTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    CURRENT_THEME: PropTypes.object.isRequired,
  };

  state = {
    anchorEl: null,
  };

  componentDidMount() {
    const caster = this.context.CURRENT_THEME;
    this.unsubscribeId = caster.subscribe(st => this.setState(st));
  }

  componentWillUnmount() {
    if (this.unsubscribeId !== null) {
      const caster = this.context.CURRENT_THEME;
      caster.unsubscribe(this.unsubscribeId);
    }
  }

  handleToggle = () => {
    const caster = this.context.CURRENT_THEME;
    const st = caster.getState();
    caster.setState({ ...st, dark: !st.dark });
  };

  handleThemeClick = event => {
    this.setState({ anchorEl: event.currentTarget });
  };

  handleThemeClose = () => {
    this.setState({ anchorEl: null });
  };

  handleThemeSelect = theme => () => {
    const caster = this.context.CURRENT_THEME;
    const st = caster.getState();
    caster.setState({ ...st, theme });
  };

  render() {
    const { classes, open, onClose } = this.props;
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
                <Switch onChange={this.handleToggle} />
              </ListItemSecondaryAction>
            </ListItem>
            <ListItem button onClick={this.handleThemeClick}>
              <ListItemIcon>
                <EyeIcon />
              </ListItemIcon>
              <ListItemText primary="Theme" secondary="Default" />
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
            <MenuItem onClick={this.handleThemeSelect("bubblegum")}>
              Bubblegum
            </MenuItem>
            <MenuItem onClick={this.handleThemeSelect("dva")}>DVA</MenuItem>
          </Menu>
        </Paper>
      </Modal>
    );
  }
}

export default withStyles(styles)(Preferences);
