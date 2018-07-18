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
            anchorEl: this._menuAnchorRef.current,
            ...props,
          })
        }
      >
        {props =>
          children({
            ref: this._menuAnchorRef,
            ...props,
          })
        }
      </ModalController>
    );
  }
}

export default MenuController;
