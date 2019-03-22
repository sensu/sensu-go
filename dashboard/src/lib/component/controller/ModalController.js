import React from "react";
import PropTypes from "prop-types";

class ModalController extends React.Component {
  static propTypes = {
    children: PropTypes.func.isRequired,
    renderModal: PropTypes.func.isRequired,
  };

  state = {
    modalOpen: false,
  };

  open = () => {
    this.setState({ modalOpen: true });
  };

  close = () => {
    this.setState({ modalOpen: false });
  };

  render() {
    const { children, renderModal } = this.props;
    const { modalOpen } = this.state;

    return (
      <React.Fragment>
        {children({ isOpen: modalOpen, open: this.open })}
        {modalOpen && renderModal({ close: this.close })}
      </React.Fragment>
    );
  }
}

export default ModalController;
