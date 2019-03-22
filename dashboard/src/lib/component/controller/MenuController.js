import React from "react";
import PropTypes from "prop-types";
import getNextId from "/lib/util/uniqueId";
import ModalController from "./ModalController";

class MenuController extends React.Component {
  static propTypes = {
    children: PropTypes.func.isRequired,
    renderMenu: PropTypes.func.isRequired,
  };

  _menuAnchorRef = React.createRef();

  constructor(props) {
    super(props);
    this._id = getNextId();
  }

  render() {
    const { children, renderMenu } = this.props;
    const idx = `menu-idx-${this._id}`;

    return (
      <ModalController
        renderModal={props =>
          renderMenu({
            anchorEl: this._menuAnchorRef.current,
            idx,
            ...props,
          })
        }
      >
        {props =>
          children({
            idx,
            ref: this._menuAnchorRef,
            ...props,
          })
        }
      </ModalController>
    );
  }
}

export default MenuController;
