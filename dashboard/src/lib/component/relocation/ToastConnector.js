import React from "react";
import PropTypes from "prop-types";

import { Consumer } from "/lib/component/relocation/Relocation";
import { TOAST } from "/lib/component/relocation/types";

class ToastConnector extends React.PureComponent {
  static propTypes = {
    children: PropTypes.func.isRequired,
  };

  render() {
    return (
      <Consumer>
        {({ createChild }) =>
          this.props.children({
            addToast: render => createChild({ render, type: TOAST }),
          })
        }
      </Consumer>
    );
  }
}

export default ToastConnector;
