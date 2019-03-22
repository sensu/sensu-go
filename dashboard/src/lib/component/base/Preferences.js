import React from "react";
import PropTypes from "prop-types";
import { connect } from "react-redux";
import { compose, withProps } from "recompose";
import withMobileDialog from "@material-ui/core/withMobileDialog";

import AppBar from "@material-ui/core/AppBar";
import BulbIcon from "@material-ui/icons/SettingsBrightness";
import CloseIcon from "@material-ui/icons/Close";
import Dialog from "@material-ui/core/Dialog";
import EyeIcon from "@material-ui/icons/RemoveRedEye";
import IconButton from "@material-ui/core/IconButton";
import List from "@material-ui/core/List";
import ListItem from "@material-ui/core/ListItem";
import ListItemIcon from "@material-ui/core/ListItemIcon";
import ListItemSecondaryAction from "@material-ui/core/ListItemSecondaryAction";
import ListItemText from "@material-ui/core/ListItemText";
import ListSubheader from "@material-ui/core/ListSubheader";
import Menu from "@material-ui/core/Menu";
import MenuItem from "@material-ui/core/MenuItem";
import MenuList from "@material-ui/core/MenuList";
import Slide from "@material-ui/core/Slide";
import Switch from "@material-ui/core/Switch";
import Toolbar from "@material-ui/core/Toolbar";
import Typography from "@material-ui/core/Typography";

const SlideUp = withProps({ direction: "up" })(Slide);

class Preferences extends React.Component {
  static propTypes = {
    dispatch: PropTypes.func.isRequired,
    fullScreen: PropTypes.bool.isRequired,
    open: PropTypes.bool.isRequired,
    onClose: PropTypes.func.isRequired,
    theme: PropTypes.shape({
      theme: PropTypes.string,
      dark: PropTypes.bool,
    }).isRequired,
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
    const { fullScreen, open, onClose, theme } = this.props;
    const { anchorEl } = this.state;

    return (
      <Dialog
        open={open}
        onClose={onClose}
        TransitionComponent={SlideUp}
        fullScreen={fullScreen}
        fullWidth
        PaperProps={{
          style: { minHeight: 400 },
        }}
      >
        <AppBar style={{ position: "relative" }}>
          <Toolbar>
            <IconButton color="inherit" onClick={onClose} aria-label="Close">
              <CloseIcon />
            </IconButton>
            <Typography variant="title" color="inherit">
              Preferences
            </Typography>
          </Toolbar>
        </AppBar>
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
          <MenuList>
            <MenuItem onClick={this.handleThemeSelect("sensu")}>
              <ListItem>
                <ListItemText primary="Default" secondary=" " />
              </ListItem>
            </MenuItem>
            <MenuItem onClick={this.handleThemeSelect("classic")}>
              <ListItem>
                <ListItemText
                  primary="Classic"
                  secondary="Vintage Sensu in apple green."
                />
              </ListItem>
            </MenuItem>
            <MenuItem onClick={this.handleThemeSelect("uchiwa")}>
              <ListItem>
                <ListItemText primary="Uchiwa" secondary="Cool in blue." />
              </ListItem>
            </MenuItem>
          </MenuList>
        </Menu>
      </Dialog>
    );
  }
}

const enhancer = compose(
  withMobileDialog({ breakpoint: "xs" }),
  connect(st => ({ theme: st.theme })),
);
export default enhancer(Preferences);
