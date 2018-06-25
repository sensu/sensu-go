import React from "react";
import PropTypes from "prop-types";

class MenuController extends React.Component {
  static propTypes = {
    children: PropTypes.func.isRequired,
    renderMenu: PropTypes.func.isRequired,
  };

  state = {
    menuOpen: false,
  };

  _menuAnchorRef = React.createRef();

  open = () => {
    this.setState({ menuOpen: true });
  };
  close = () => {
    this.setState({ menuOpen: false });
  };

  render() {
    const { children, renderMenu } = this.props;
    const { menuOpen } = this.state;

    return (
      <React.Fragment>
        {children({
          open: this.open,
          ref: this._menuAnchorRef,
        })}
        {menuOpen &&
          renderMenu({
            anchorEl: this._menuAnchorRef.current,
            close: this.close,
          })}
      </React.Fragment>
    );
  }
}

export default MenuController;
