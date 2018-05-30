import React from "react";
import PropTypes from "prop-types";
import ConfirmAction from "/components/util/ConfirmAction";
import Dialog from "./Dialog";

class ConfirmDelete extends React.Component {
  static propTypes = {
    onSubmit: PropTypes.func.isRequired,
    children: PropTypes.func.isRequired,
  };

  render() {
    const { onSubmit, children, ...props } = this.props;
    return (
      <ConfirmAction>
        {({ isOpen, open, close }) => (
          <React.Fragment>
            <Dialog
              {...props}
              open={isOpen}
              onClose={close}
              onConfirm={ev => {
                onSubmit(ev);
                close();
              }}
            />
            {children({ open, close })}
          </React.Fragment>
        )}
      </ConfirmAction>
    );
  }
}

export default ConfirmDelete;
