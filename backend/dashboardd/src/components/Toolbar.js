import React from "react";
import PropTypes from "prop-types";
import AppBar from "material-ui/AppBar";

import MaterialToolbar from "material-ui/Toolbar";
import Typography from "material-ui/Typography";
import Button from "material-ui/Button";
import IconButton from "material-ui/IconButton";
import MenuIcon from "material-ui-icons/Menu";

function Toolbar({ toggleToolbar }) {
  return (
    <AppBar position="static">
      <MaterialToolbar>
        <IconButton onClick={toggleToolbar} color="contrast" aria-label="Menu">
          <MenuIcon />
        </IconButton>
        <Typography type="title" color="inherit">
          Sensu
        </Typography>
        <Button color="contrast">Login</Button>
      </MaterialToolbar>
    </AppBar>
  );
}

Toolbar.propTypes = {
  toggleToolbar: PropTypes.func.isRequired,
};

export default Toolbar;
