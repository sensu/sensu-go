import React from "react";
import PropTypes from "prop-types";
import ModalController from "./ModalController";

class MenuController extends React.Component {
  static propTypes = {
    children: PropTypes.func.isRequired,
    renderMenu: PropTypes.func.isRequired,
  };

  _menuAnchorRef = React.createRef();

  render() {
    const { children, renderMenu } = this.props;

    return (
      <ModalController
        renderModal={props =>
          renderMenu({
            ref: this._menuAnchorRef,
            ...props,
          })
        }
      >
        {props =>
          children({
            anchorEl: this.menuAnchorRef.current,
            ...props,
          })
        }
      </ModalController>
    );
  }
}

export default MenuController;
