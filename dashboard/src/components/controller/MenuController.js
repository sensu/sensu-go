import React from "react";
import PropTypes from "prop-types";
import ModalController from "./ModalController";

let id = 0;
const getNextId = () => {
  id += 1;
  return id;
};

class MenuController extends React.Component {
  static propTypes = {
    children: PropTypes.func.isRequired,
    renderMenu: PropTypes.func.isRequired,
  };

  constructor(props) {
    super(props);
    this._id = getNextId();
  }

  _menuAnchorRef = React.createRef();

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
